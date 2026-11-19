package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shiftschedule/internal/domain"
)

const personnelSelect = `SELECT id, name, email, groups, is_admin FROM personnel`

func scanPersonnel(row pgx.CollectableRow) (domain.Personnel, error) {
	var p domain.Personnel
	err := row.Scan(&p.ID, &p.Name, &p.Email, &p.Groups, &p.IsAdmin)
	return p, err
}

// nonNil makes sure a nil slice is stored as an empty array instead of NULL.
func nonNil(groups []string) []string {
	if groups == nil {
		return []string{}
	}
	return groups
}

func (s *Store) PersonByEmail(ctx context.Context, email string) (domain.Personnel, error) {
	rows, _ := s.pool.Query(ctx, personnelSelect+` WHERE lower(email) = lower($1)`, email)
	p, err := pgx.CollectExactlyOneRow(rows, scanPersonnel)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Personnel{}, domain.ErrNotFound
	}
	return p, err
}

func (s *Store) ListPersonnel(ctx context.Context) ([]domain.Personnel, error) {
	rows, _ := s.pool.Query(ctx, personnelSelect+` ORDER BY name`)
	return pgx.CollectRows(rows, scanPersonnel)
}

func (s *Store) CreatePersonnel(ctx context.Context, name, email string, groups []string, isAdmin bool) (domain.Personnel, error) {
	rows, _ := s.pool.Query(ctx, `
		INSERT INTO personnel (name, email, groups, is_admin) VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, groups, is_admin`, name, email, nonNil(groups), isAdmin)
	p, err := pgx.CollectExactlyOneRow(rows, scanPersonnel)
	if err != nil {
		return domain.Personnel{}, mapError(err)
	}
	return p, s.SyncMembership(ctx)
}

// UpsertPersonnel creates or updates a person identified by email.
func (s *Store) UpsertPersonnel(ctx context.Context, name, email string, groups []string, isAdmin bool) (domain.Personnel, error) {
	rows, _ := s.pool.Query(ctx, `
		INSERT INTO personnel (name, email, groups, is_admin) VALUES ($1, $2, $3, $4)
		ON CONFLICT (lower(email)) DO UPDATE SET name = EXCLUDED.name, groups = EXCLUDED.groups, is_admin = EXCLUDED.is_admin
		RETURNING id, name, email, groups, is_admin`, name, email, nonNil(groups), isAdmin)
	p, err := pgx.CollectExactlyOneRow(rows, scanPersonnel)
	if err != nil {
		return domain.Personnel{}, mapError(err)
	}
	return p, s.SyncMembership(ctx)
}

// SetGroups replaces a person's groups and updates the team memberships that follow from them.
func (s *Store) SetGroups(ctx context.Context, personID int64, groups []string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE personnel SET groups = $2 WHERE id = $1`, personID, nonNil(groups))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return s.SyncMembership(ctx)
}

// UpsertTeam creates or updates a team. An empty group means membership is managed with AddTeamMember instead of OIDC groups.
func (s *Store) UpsertTeam(ctx context.Context, name, group string, managerID *int64) (domain.Team, error) {
	rows, _ := s.pool.Query(ctx, `
		INSERT INTO team (name, group_name, manager_id) VALUES ($1, NULLIF($2, ''), $3)
		ON CONFLICT (name) DO UPDATE SET group_name = EXCLUDED.group_name, manager_id = EXCLUDED.manager_id
		RETURNING id, name, COALESCE(group_name, ''), manager_id`, name, group, managerID)
	team, err := pgx.CollectExactlyOneRow(rows, func(row pgx.CollectableRow) (domain.Team, error) {
		var t domain.Team
		err := row.Scan(&t.ID, &t.Name, &t.Group, &t.ManagerID)
		return t, err
	})
	if err != nil {
		return domain.Team{}, mapError(err)
	}
	return team, s.SyncMembership(ctx)
}

// AddTeamMember adds a member to a team. Teams with a group have their members overwritten by SyncMembership.
func (s *Store) AddTeamMember(ctx context.Context, teamID, personID int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO team_member (team_id, person_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, teamID, personID)
	return mapError(err)
}

// SyncMembership makes team_member match personnel.groups for every team that has a group.
// Call it whenever groups change, for example after an OIDC login.
func (s *Store) SyncMembership(ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		DELETE FROM team_member tm USING team t
		WHERE tm.team_id = t.id AND t.group_name IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM personnel p WHERE p.id = tm.person_id AND t.group_name = ANY (p.groups))`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO team_member (team_id, person_id)
		SELECT t.id, p.id FROM team t JOIN personnel p ON t.group_name = ANY (p.groups)
		ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// UpsertPosition creates the position if missing and returns it.
func (s *Store) UpsertPosition(ctx context.Context, teamID int64, name string, kind domain.ShiftKind) (domain.Position, error) {
	if !kind.Valid() {
		return domain.Position{}, fmt.Errorf("%w: unknown shift kind %q", domain.ErrInvalid, kind)
	}

	rows, _ := s.pool.Query(ctx, `
		WITH p AS (
			INSERT INTO position (team_id, name, kind) VALUES ($1, $2, $3)
			ON CONFLICT (team_id, name) DO UPDATE SET kind = EXCLUDED.kind
			RETURNING id, team_id, name, kind
		)
		SELECT p.id, p.team_id, t.name, t.manager_id, p.name, p.kind FROM p JOIN team t ON t.id = p.team_id`, teamID, name, string(kind))
	pos, err := pgx.CollectExactlyOneRow(rows, scanPosition)
	return pos, mapError(err)
}

const positionSelect = `
	SELECT p.id, p.team_id, t.name, t.manager_id, p.name, p.kind
	FROM position p JOIN team t ON t.id = p.team_id`

func scanPosition(row pgx.CollectableRow) (domain.Position, error) {
	var p domain.Position
	var kind string
	err := row.Scan(&p.ID, &p.TeamID, &p.TeamName, &p.ManagerID, &p.Name, &kind)
	p.Kind = domain.ShiftKind(kind)
	return p, err
}

// ListPositions returns all positions, optionally only of one kind.
func (s *Store) ListPositions(ctx context.Context, kind *domain.ShiftKind) ([]domain.Position, error) {
	var k *string
	if kind != nil {
		v := string(*kind)
		k = &v
	}
	rows, _ := s.pool.Query(ctx, positionSelect+` WHERE ($1::text IS NULL OR p.kind = $1) ORDER BY t.name, p.kind, p.name`, k)
	return pgx.CollectRows(rows, scanPosition)
}

// PositionsForPerson returns the positions of the teams the person is a member of.
func (s *Store) PositionsForPerson(ctx context.Context, personID int64) ([]domain.Position, error) {
	rows, _ := s.pool.Query(ctx, positionSelect+`
		JOIN team_member tm ON tm.team_id = p.team_id
		WHERE tm.person_id = $1
		ORDER BY t.name, p.kind, p.name`, personID)
	return pgx.CollectRows(rows, scanPosition)
}

func (s *Store) PositionByID(ctx context.Context, id int64) (domain.Position, error) {
	rows, _ := s.pool.Query(ctx, positionSelect+` WHERE p.id = $1`, id)
	pos, err := pgx.CollectExactlyOneRow(rows, scanPosition)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Position{}, domain.ErrNotFound
	}
	return pos, err
}

// IsTeamMember reports whether the person is a member of the team that owns the position.
func (s *Store) IsTeamMember(ctx context.Context, positionID, personID int64) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM position p JOIN team_member tm ON tm.team_id = p.team_id
			WHERE p.id = $1 AND tm.person_id = $2
		)`, positionID, personID).Scan(&ok)
	return ok, err
}

const assignmentSelect = `
	SELECT a.id, a.position_id, a.starts_at, a.ends_at, a.assignee_id, p.name
	FROM assignment a LEFT JOIN personnel p ON p.id = a.assignee_id`

func scanAssignment(row pgx.CollectableRow) (domain.Assignment, error) {
	var a domain.Assignment
	var assigneeID *int64
	var assigneeName *string
	if err := row.Scan(&a.ID, &a.PositionID, &a.StartsAt, &a.EndsAt, &assigneeID, &assigneeName); err != nil {
		return a, err
	}
	if assigneeID != nil && assigneeName != nil {
		a.Assignee = &domain.PersonRef{ID: *assigneeID, Name: *assigneeName}
	}
	return a, nil
}

// AssignmentsAround returns, per position, the assignment that ended last before now, the one running at now,
// and the one that starts first after now. Missing ones are left out.
func (s *Store) AssignmentsAround(ctx context.Context, positionIDs []int64, now time.Time) ([]domain.Assignment, error) {
	rows, _ := s.pool.Query(ctx, `
		(`+assignmentSelectDistinct+` WHERE a.position_id = ANY($1) AND a.ends_at <= $2 ORDER BY a.position_id, a.ends_at DESC)
		UNION ALL
		(`+assignmentSelectDistinct+` WHERE a.position_id = ANY($1) AND a.starts_at <= $2 AND a.ends_at > $2 ORDER BY a.position_id, a.starts_at DESC)
		UNION ALL
		(`+assignmentSelectDistinct+` WHERE a.position_id = ANY($1) AND a.starts_at > $2 ORDER BY a.position_id, a.starts_at)`,
		positionIDs, now)
	return s.withCovers(ctx, rows)
}

const assignmentSelectDistinct = `
	SELECT DISTINCT ON (a.position_id) a.id, a.position_id, a.starts_at, a.ends_at, a.assignee_id, p.name
	FROM assignment a LEFT JOIN personnel p ON p.id = a.assignee_id`

// NextAssignmentsFor returns, per position, the first assignment of the person that starts after the given time.
func (s *Store) NextAssignmentsFor(ctx context.Context, personID int64, positionIDs []int64, after time.Time) ([]domain.Assignment, error) {
	rows, _ := s.pool.Query(ctx, assignmentSelectDistinct+`
		WHERE a.assignee_id = $1 AND a.position_id = ANY($2) AND a.starts_at > $3
		ORDER BY a.position_id, a.starts_at`, personID, positionIDs, after)
	return s.withCovers(ctx, rows)
}

// CoverOnAssignment is a cover together with where it applies.
type CoverOnAssignment struct {
	AssignmentStart time.Time
	Cover           domain.Cover
	PositionID      int64
}

// NextCoversFor returns, per position, the person's first cover that has not ended at the given time.
func (s *Store) NextCoversFor(ctx context.Context, personID int64, positionIDs []int64, now time.Time) ([]CoverOnAssignment, error) {
	rows, _ := s.pool.Query(ctx, `
		SELECT DISTINCT ON (a.position_id)
			a.position_id, a.starts_at, c.id, c.assignment_id, c.starts_at, c.ends_at, c.reason, c.covered_by, p.name
		FROM cover c
		JOIN assignment a ON a.id = c.assignment_id
		JOIN personnel p ON p.id = c.covered_by
		WHERE c.covered_by = $1 AND a.position_id = ANY($2) AND c.ends_at > $3
		ORDER BY a.position_id, c.starts_at`, personID, positionIDs, now)
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (CoverOnAssignment, error) {
		var c CoverOnAssignment
		err := row.Scan(&c.PositionID, &c.AssignmentStart, &c.Cover.ID, &c.Cover.AssignmentID, &c.Cover.StartsAt, &c.Cover.EndsAt,
			&c.Cover.Reason, &c.Cover.CoveredBy.ID, &c.Cover.CoveredBy.Name)
		return c, err
	})
}

// AssignmentAt returns the assignment of the position that is running at the given time.
func (s *Store) AssignmentAt(ctx context.Context, positionID int64, at time.Time) (domain.Assignment, error) {
	rows, _ := s.pool.Query(ctx, assignmentSelect+` WHERE a.position_id = $1 AND a.starts_at <= $2 AND a.ends_at > $2`, positionID, at)
	return s.oneAssignment(ctx, rows)
}

func (s *Store) AssignmentByCoverID(ctx context.Context, coverID int64) (domain.Assignment, error) {
	rows, _ := s.pool.Query(ctx, assignmentSelect+` JOIN cover c ON c.assignment_id = a.id WHERE c.id = $1`, coverID)
	return s.oneAssignment(ctx, rows)
}

func (s *Store) oneAssignment(ctx context.Context, rows pgx.Rows) (domain.Assignment, error) {
	list, err := s.withCovers(ctx, rows)
	if err != nil {
		return domain.Assignment{}, err
	}
	if len(list) == 0 {
		return domain.Assignment{}, domain.ErrNotFound
	}
	return list[0], nil
}

// UpsertAssignment sets who holds a position for the period starting at startsAt. A nil assigneeID leaves it unfilled.
// It returns domain.ErrConflict if the period overlaps a different assignment of the position.
func (s *Store) UpsertAssignment(
	ctx context.Context, positionID int64, startsAt, endsAt time.Time, assigneeID *int64, createdBy int64,
) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO assignment (position_id, starts_at, ends_at, assignee_id, created_by) VALUES ($1, $2, $3, $4, NULLIF($5, 0))
		ON CONFLICT (position_id, starts_at) DO UPDATE
		SET ends_at = EXCLUDED.ends_at, assignee_id = EXCLUDED.assignee_id, updated_at = now()`,
		positionID, startsAt, endsAt, assigneeID, createdBy)
	return mapError(err)
}

// withCovers collects assignments from rows and attaches their covers.
func (s *Store) withCovers(ctx context.Context, rows pgx.Rows) ([]domain.Assignment, error) {
	assignments, err := pgx.CollectRows(rows, scanAssignment)
	if err != nil || len(assignments) == 0 {
		return assignments, err
	}

	ids := make([]int64, len(assignments))
	index := make(map[int64]int, len(assignments))
	for i, a := range assignments {
		ids[i] = a.ID
		index[a.ID] = i
	}

	coverRows, _ := s.pool.Query(ctx, `
		SELECT c.id, c.assignment_id, c.starts_at, c.ends_at, c.reason, c.covered_by, p.name
		FROM cover c JOIN personnel p ON p.id = c.covered_by
		WHERE c.assignment_id = ANY($1)
		ORDER BY c.starts_at`, ids)
	covers, err := pgx.CollectRows(coverRows, scanCover)
	if err != nil {
		return nil, err
	}

	for _, c := range covers {
		i := index[c.AssignmentID]
		assignments[i].Covers = append(assignments[i].Covers, c)
	}
	return assignments, nil
}

func scanCover(row pgx.CollectableRow) (domain.Cover, error) {
	var c domain.Cover
	err := row.Scan(&c.ID, &c.AssignmentID, &c.StartsAt, &c.EndsAt, &c.Reason, &c.CoveredBy.ID, &c.CoveredBy.Name)
	return c, err
}

// CreateCover records that coveredBy takes over an assignment between startsAt and endsAt.
// It returns domain.ErrConflict if the period overlaps an existing cover.
func (s *Store) CreateCover(
	ctx context.Context, assignmentID, coveredBy int64, startsAt, endsAt time.Time, reason string, createdBy int64,
) (domain.Cover, error) {
	rows, _ := s.pool.Query(ctx, `
		WITH c AS (
			INSERT INTO cover (assignment_id, covered_by, starts_at, ends_at, reason, created_by) VALUES ($1, $2, $3, $4, $5, NULLIF($6, 0))
			RETURNING id, assignment_id, starts_at, ends_at, reason, covered_by
		)
		SELECT c.id, c.assignment_id, c.starts_at, c.ends_at, c.reason, c.covered_by, p.name
		FROM c JOIN personnel p ON p.id = c.covered_by`, assignmentID, coveredBy, startsAt, endsAt, reason, createdBy)
	cover, err := pgx.CollectExactlyOneRow(rows, scanCover)
	return cover, mapError(err)
}

func (s *Store) DeleteCover(ctx context.Context, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM cover WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RedDayOverrides returns the admin overrides keyed by date (UTC midnight).
func (s *Store) RedDayOverrides(ctx context.Context) (map[time.Time]bool, error) {
	rows, _ := s.pool.Query(ctx, `SELECT day, is_red_day FROM red_day_override`)
	overrides := map[time.Time]bool{}
	var day time.Time
	var isRed bool
	_, err := pgx.ForEachRow(rows, []any{&day, &isRed}, func() error {
		overrides[day] = isRed
		return nil
	})
	return overrides, err
}

func (s *Store) SetRedDayOverride(ctx context.Context, day time.Time, name string, isRedDay bool) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO red_day_override (day, name, is_red_day) VALUES ($1, $2, $3)
		ON CONFLICT (day) DO UPDATE SET name = EXCLUDED.name, is_red_day = EXCLUDED.is_red_day`, day, name, isRedDay)
	return err
}

func (s *Store) DeleteRedDayOverride(ctx context.Context, day time.Time) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM red_day_override WHERE day = $1`, day)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// RotationInputs returns every position with its team members (alphabetical by name) and the last assignee that started
// before the given time, grouped by team.
func (s *Store) RotationInputs(ctx context.Context, before time.Time) (map[int64][]domain.RotationPosition, error) {
	rows, _ := s.pool.Query(ctx, `
		SELECT p.team_id, p.id, p.kind,
			COALESCE(ARRAY(
				SELECT tm.person_id FROM team_member tm JOIN personnel pe ON pe.id = tm.person_id
				WHERE tm.team_id = p.team_id ORDER BY lower(pe.name), pe.id), '{}'),
			COALESCE((SELECT a.assignee_id FROM assignment a
				WHERE a.position_id = p.id AND a.starts_at < $1 AND a.assignee_id IS NOT NULL
				ORDER BY a.starts_at DESC LIMIT 1), 0)
		FROM position p
		ORDER BY p.team_id, p.id`, before)

	type row struct {
		pos    domain.RotationPosition
		teamID int64
	}
	list, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (row, error) {
		var x row
		var kind string
		err := r.Scan(&x.teamID, &x.pos.PositionID, &kind, &x.pos.Pool, &x.pos.Last)
		x.pos.Kind = domain.ShiftKind(kind)
		return x, err
	})
	if err != nil {
		return nil, err
	}

	byTeam := map[int64][]domain.RotationPosition{}
	for _, x := range list {
		byTeam[x.teamID] = append(byTeam[x.teamID], x.pos)
	}
	return byTeam, nil
}

// InsertPlanned stores generated assignments. Existing unfilled assignments at the same start are filled in,
// and periods that already have an assignee are left alone. It returns how many assignments were created or filled in.
func (s *Store) InsertPlanned(ctx context.Context, planned []domain.PlannedSlot) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	changed := 0
	for _, p := range planned {
		tag, err := tx.Exec(ctx, `
			UPDATE assignment SET assignee_id = $3, updated_at = now()
			WHERE position_id = $1 AND starts_at = $2 AND assignee_id IS NULL`, p.PositionID, p.StartsAt, p.AssigneeID)
		if err != nil {
			return 0, mapError(err)
		}
		if tag.RowsAffected() == 0 {
			tag, err = tx.Exec(ctx, `
				INSERT INTO assignment (position_id, starts_at, ends_at, assignee_id) VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING`, p.PositionID, p.StartsAt, p.EndsAt, p.AssigneeID)
			if err != nil {
				return 0, mapError(err)
			}
		}
		changed += int(tag.RowsAffected())
	}

	return changed, tx.Commit(ctx)
}
