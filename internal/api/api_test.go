package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/shiftschedule/internal/api"
	"github.com/shiftschedule/internal/auth"
	"github.com/shiftschedule/internal/domain"
	"github.com/shiftschedule/internal/service"
	"github.com/shiftschedule/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	admin    = "admin@example.com"
	ola      = "ola@example.com"
	kari     = "kari@example.com"
	per      = "per@example.com"
	outsider = "outsider@example.com"
)

type fixture struct {
	server  *httptest.Server
	ids     map[string]int64
	dayPos  int64
	callPos int64
}

type envelope struct {
	Message string          `json:"message"`
	Body    json.RawMessage `json:"body"`
	Code    int             `json:"code"`
}

// newFixture seeds one team (ola, kari, per; managed by kari) with a day shift and an on-call position, with assignments around
// Wednesday 2026-11-18 08:34 (Europe/Oslo). The current week starts Monday 2026-11-16.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureWithAuth(t, &auth.Dev{})
}

func newFixtureWithAuth(t *testing.T, authenticator auth.Authenticator) *fixture {
	t.Helper()
	ctx := context.Background()
	store := testutil.NewStore(t)

	loc, err := time.LoadLocation("Europe/Oslo")
	require.NoError(t, err)
	now := time.Date(2026, 11, 18, 8, 34, 0, 0, loc)

	f := &fixture{ids: map[string]int64{}}
	for _, p := range []struct {
		email   string
		groups  []string
		isAdmin bool
	}{{admin, nil, true}, {ola, []string{"alpha"}, false}, {kari, []string{"alpha"}, false}, {per, []string{"alpha"}, false}, {outsider, nil, false}} {
		saved, err := store.UpsertPersonnel(ctx, p.email, p.email, p.groups, p.isAdmin)
		require.NoError(t, err)
		f.ids[p.email] = saved.ID
	}

	manager := f.ids[kari]
	team, err := store.UpsertTeam(ctx, "Alpha", "alpha", &manager)
	require.NoError(t, err)
	day, err := store.UpsertPosition(ctx, team.ID, "Day", domain.KindDayShift)
	require.NoError(t, err)
	call, err := store.UpsertPosition(ctx, team.ID, "On-call", domain.KindOnCall)
	require.NoError(t, err)
	f.dayPos, f.callPos = day.ID, call.ID

	cal := domain.Calendar{Loc: loc}
	assign := func(position int64, week string, who string) {
		monday, err := time.Parse("2006-01-02", week)
		require.NoError(t, err)
		start, end := cal.WeekBounds(monday)
		assignee := f.ids[who]
		require.NoError(t, store.UpsertAssignment(ctx, position, start, end, &assignee, f.ids[admin]))
	}
	for week, who := range map[string]string{"2026-11-09": kari, "2026-11-16": ola, "2026-11-23": kari, "2026-11-30": ola} {
		assign(day.ID, week, who)
	}
	for week, who := range map[string]string{"2026-11-09": ola, "2026-11-16": kari, "2026-11-23": per, "2026-11-30": ola} {
		assign(call.ID, week, who)
	}

	svc := service.New(store, loc, func() time.Time { return now })
	handler := api.NewHandler(zerolog.Nop(), svc, authenticator)
	f.server = httptest.NewServer(handler.Routes())
	t.Cleanup(f.server.Close)

	return f
}

func (f *fixture) call(t *testing.T, user, method, path string, body any) (int, envelope) {
	t.Helper()

	var reader bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&reader).Encode(body))
	}
	req, err := http.NewRequestWithContext(context.Background(), method, f.server.URL+path, &reader)
	require.NoError(t, err)
	if user != "" {
		req.Header.Set(auth.DevHeader, user)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var env envelope
	if resp.StatusCode != http.StatusNoContent {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&env))
	}
	return resp.StatusCode, env
}

type personJSON struct {
	Name string `json:"name"`
}

type weekJSON struct {
	Assignee  *personJSON `json:"assignee"`
	Holder    *personJSON `json:"holder"`
	WeekStart string      `json:"week_start"`
	Covers    []struct {
		Ends time.Time `json:"ends_at"`
		ID   int64     `json:"id"`
		IsMe bool      `json:"is_me"`
	} `json:"covers"`
	IsMe bool `json:"is_me"`
}

type scheduleJSON struct {
	Previous *weekJSON `json:"previous"`
	Current  *weekJSON `json:"current"`
	Next     *weekJSON `json:"next"`
	MyNext   *weekJSON `json:"my_next"`
	Position struct {
		Name string `json:"name"`
	} `json:"position"`
}

func TestMyScheduleShowsPreviousCurrentNextAndMyNext(t *testing.T) {
	f := newFixture(t)

	status, env := f.call(t, ola, http.MethodGet, "/v1/me/schedule", nil)
	require.Equal(t, http.StatusOK, status)

	var schedules []scheduleJSON
	require.NoError(t, json.Unmarshal(env.Body, &schedules))
	require.Len(t, schedules, 2)

	day := schedules[0]
	assert.Equal(t, "Day", day.Position.Name)
	assert.Equal(t, "kari@example.com", day.Previous.Assignee.Name)
	assert.False(t, day.Previous.IsMe)
	assert.Equal(t, "2026-11-16", day.Current.WeekStart)
	assert.True(t, day.Current.IsMe, "ola holds the current day shift and should be highlighted")
	assert.Equal(t, "kari@example.com", day.Next.Assignee.Name)
	assert.Equal(t, "2026-11-30", day.MyNext.WeekStart)

	onCall := schedules[1]
	assert.True(t, onCall.Previous.IsMe)
	assert.False(t, onCall.Current.IsMe)
	assert.Equal(t, "2026-11-30", onCall.MyNext.WeekStart)
}

func TestMyScheduleIsEmptyWithoutPositions(t *testing.T) {
	f := newFixture(t)

	status, env := f.call(t, outsider, http.MethodGet, "/v1/me/schedule", nil)

	require.Equal(t, http.StatusOK, status)
	assert.JSONEq(t, `[]`, string(env.Body))
}

func TestOverviewSplitsDayShiftAndOnCall(t *testing.T) {
	f := newFixture(t)

	status, env := f.call(t, per, http.MethodGet, "/v1/schedules/overview", nil)
	require.Equal(t, http.StatusOK, status)

	var overview struct {
		DayShift []scheduleJSON `json:"day_shift"`
		OnCall   []scheduleJSON `json:"on_call"`
	}
	require.NoError(t, json.Unmarshal(env.Body, &overview))
	require.Len(t, overview.DayShift, 1)
	require.Len(t, overview.OnCall, 1)

	assert.Nil(t, overview.DayShift[0].MyNext, "per is not in the day shift")
	assert.Equal(t, "2026-11-23", overview.OnCall[0].MyNext.WeekStart)

	status, _ = f.call(t, per, http.MethodGet, "/v1/schedules/overview?kind=nonsense", nil)
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestCoverRestOfDayChangesHolderAndKeepsOriginalVisible(t *testing.T) {
	f := newFixture(t)
	cover := map[string]any{"position_id": f.callPos, "covered_by": f.ids[per], "scope": "rest_of_day"}

	status, _ := f.call(t, ola, http.MethodPost, "/v1/covers", cover)
	assert.Equal(t, http.StatusForbidden, status, "ola is not the assignee")

	status, env := f.call(t, kari, http.MethodPost, "/v1/covers", cover)
	require.Equal(t, http.StatusCreated, status, env.Message)
	var created struct {
		Ends time.Time `json:"ends_at"`
		ID   int64     `json:"id"`
	}
	require.NoError(t, json.Unmarshal(env.Body, &created))
	assert.Equal(t, "2026-11-19T07:00:00Z", created.Ends.UTC().Format(time.RFC3339), "on-call hands over at 08:00 Oslo")

	status, env = f.call(t, per, http.MethodGet, "/v1/me/schedule", nil)
	require.Equal(t, http.StatusOK, status)
	var schedules []scheduleJSON
	require.NoError(t, json.Unmarshal(env.Body, &schedules))
	require.Len(t, schedules, 2)
	current := schedules[1].Current
	assert.Equal(t, "kari@example.com", current.Assignee.Name, "the original stays visible")
	assert.Equal(t, "per@example.com", current.Holder.Name)
	assert.True(t, current.IsMe)
	require.Len(t, current.Covers, 1)
	assert.True(t, current.Covers[0].IsMe)

	status, env = f.call(t, kari, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.callPos, "covered_by": f.ids[ola], "scope": "whole_week"})
	assert.Equal(t, http.StatusConflict, status, "overlaps the existing cover: %s", env.Message)

	status, _ = f.call(t, per, http.MethodDelete, "/v1/covers/"+strconv.FormatInt(created.ID, 10), nil)
	assert.Equal(t, http.StatusNoContent, status, "the person covering may give it back")
}

func TestCoverRules(t *testing.T) {
	f := newFixture(t)

	status, _ := f.call(t, kari, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.callPos, "covered_by": f.ids[outsider], "scope": "whole_week"})
	assert.Equal(t, http.StatusBadRequest, status, "cover must be in the pool")

	status, _ = f.call(t, kari, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.callPos, "covered_by": f.ids[kari], "scope": "whole_week"})
	assert.Equal(t, http.StatusBadRequest, status, "cannot cover yourself")

	status, _ = f.call(t, kari, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.callPos, "covered_by": f.ids[per], "scope": "forever"})
	assert.Equal(t, http.StatusBadRequest, status, "unknown scope")

	status, _ = f.call(t, ola, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.callPos, "covered_by": f.ids[per], "scope": "whole_week", "week_start": "2026-11-09"})
	assert.Equal(t, http.StatusConflict, status, "last week is over")

	status, _ = f.call(t, per, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.callPos, "covered_by": f.ids[kari], "scope": "rest_of_day", "week_start": "2026-11-23"})
	assert.Equal(t, http.StatusBadRequest, status, "rest of day only applies to the current week")

	status, _ = f.call(t, ola, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.dayPos, "covered_by": f.ids[kari], "scope": "rest_of_day"})
	assert.Equal(t, http.StatusCreated, status, "ola holds the day shift this week and it is before 15:30")
}

func TestAuthAndAdmin(t *testing.T) {
	f := newFixture(t)

	status, _ := f.call(t, "", http.MethodGet, "/v1/me", nil)
	assert.Equal(t, http.StatusUnauthorized, status)

	status, _ = f.call(t, "nobody@example.com", http.MethodGet, "/v1/me", nil)
	assert.Equal(t, http.StatusForbidden, status, "not in personnel")

	status, _ = f.call(t, kari, http.MethodGet, "/v1/admin/personnel", nil)
	assert.Equal(t, http.StatusForbidden, status)

	status, _ = f.call(t, admin, http.MethodGet, "/v1/admin/personnel", nil)
	assert.Equal(t, http.StatusOK, status)

}

func TestGenerateRotationFillsYearOnceAndLeavesExistingAlone(t *testing.T) {
	f := newFixture(t)

	status, env := f.call(t, admin, http.MethodPost, "/v1/admin/rotations", map[string]any{"year": 2027})
	require.Equal(t, http.StatusCreated, status, env.Message)
	assert.JSONEq(t, `{"created": 104}`, string(env.Body))

	_, env = f.call(t, admin, http.MethodPost, "/v1/admin/rotations", map[string]any{"year": 2027})
	assert.JSONEq(t, `{"created": 0}`, string(env.Body))
}

func TestRedDayOverrideAffectsDayShiftCover(t *testing.T) {
	f := newFixture(t)

	status, _ := f.call(t, admin, http.MethodPut, "/v1/admin/red-days/2026-11-18", map[string]any{"name": "Company day", "is_red_day": true})
	require.Equal(t, http.StatusNoContent, status)

	status, env := f.call(t, ola, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.dayPos, "covered_by": f.ids[kari], "scope": "rest_of_day"})
	assert.Equal(t, http.StatusConflict, status, "no day shift on a red day: %s", env.Message)

	status, _ = f.call(t, admin, http.MethodDelete, "/v1/admin/red-days/2026-11-18", nil)
	assert.Equal(t, http.StatusNoContent, status)
}

func TestAssignmentsAreManagedByAdminsAndTeamManager(t *testing.T) {
	f := newFixture(t)
	set := func(user string, assignee any, week string) int {
		status, _ := f.call(t, user, http.MethodPut, "/v1/assignments",
			map[string]any{"position_id": f.dayPos, "assignee_id": assignee, "week_start": week})
		return status
	}

	assert.Equal(t, http.StatusForbidden, set(ola, f.ids[kari], "2026-12-07"), "only admin and manager")
	assert.Equal(t, http.StatusNoContent, set(kari, f.ids[per], "2026-12-07"), "kari manages the team")
	assert.Equal(t, http.StatusNoContent, set(admin, f.ids[ola], "2026-12-14"))
	assert.Equal(t, http.StatusBadRequest, set(kari, f.ids[outsider], "2026-12-21"), "outsider is not in the team")
	assert.Equal(t, http.StatusBadRequest, set(kari, f.ids[kari], "2026-12-08"), "not a monday")

	// The manager may also arrange a cover on someone else's week.
	status, _ := f.call(t, kari, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.dayPos, "covered_by": f.ids[per], "scope": "whole_week"})
	assert.Equal(t, http.StatusCreated, status)
}

func TestUnfilledWeekIsShownEmptyAndCannotBeCovered(t *testing.T) {
	f := newFixture(t)

	status, _ := f.call(t, kari, http.MethodPut, "/v1/assignments",
		map[string]any{"position_id": f.dayPos, "assignee_id": nil, "week_start": "2026-11-16"})
	require.Equal(t, http.StatusNoContent, status)

	status, env := f.call(t, per, http.MethodGet, "/v1/me/schedule", nil)
	require.Equal(t, http.StatusOK, status)
	var schedules []scheduleJSON
	require.NoError(t, json.Unmarshal(env.Body, &schedules))
	require.NotNil(t, schedules[0].Current)
	assert.Nil(t, schedules[0].Current.Assignee)
	assert.Nil(t, schedules[0].Current.Holder)
	assert.False(t, schedules[0].Current.IsMe)

	status, _ = f.call(t, kari, http.MethodPost, "/v1/covers",
		map[string]any{"position_id": f.dayPos, "covered_by": f.ids[per], "scope": "whole_week"})
	assert.Equal(t, http.StatusBadRequest, status, "nobody to cover for")
}

func TestTeamMembershipFollowsGroups(t *testing.T) {
	f := newFixture(t)

	status, env := f.call(t, admin, http.MethodPost, "/v1/admin/personnel",
		map[string]any{"name": "Nina", "email": "nina@example.com", "groups": []string{"alpha"}})
	require.Equal(t, http.StatusCreated, status, env.Message)

	status, env = f.call(t, "nina@example.com", http.MethodGet, "/v1/me/schedule", nil)
	require.Equal(t, http.StatusOK, status)
	var schedules []scheduleJSON
	require.NoError(t, json.Unmarshal(env.Body, &schedules))
	assert.Len(t, schedules, 2, "nina is in group alpha and sees the team's positions")
}

// stubIdP stands in for an identity provider: the dev header picks the user and the stub supplies the claims.
type stubIDP struct {
	err    error
	claims map[string]auth.Identity
}

func (s *stubIDP) Authenticate(r *http.Request) (auth.Identity, error) {
	if s.err != nil {
		return auth.Identity{}, s.err
	}
	identity, ok := s.claims[r.Header.Get(auth.DevHeader)]
	if !ok {
		return auth.Identity{}, auth.ErrUnauthenticated
	}
	return identity, nil
}

func TestLoginSyncsGroupsFromIdentityProvider(t *testing.T) {
	idp := stubIDP{claims: map[string]auth.Identity{
		ola:  {Email: ola, Groups: []string{"alpha"}, HasGroups: true},
		per:  {Email: per, Groups: []string{}, HasGroups: true}, // removed from alpha at the provider
		kari: {Email: kari},                                     // provider sends no groups claim
	}}
	f := newFixtureWithAuth(t, &idp)

	mySchedules := func(user string) int {
		status, env := f.call(t, user, http.MethodGet, "/v1/me/schedule", nil)
		require.Equal(t, http.StatusOK, status, env.Message)
		var schedules []scheduleJSON
		require.NoError(t, json.Unmarshal(env.Body, &schedules))
		return len(schedules)
	}

	assert.Equal(t, 2, mySchedules(ola))
	assert.Equal(t, 0, mySchedules(per), "empty groups claim removes the membership")
	assert.Equal(t, 2, mySchedules(kari), "a missing groups claim leaves memberships alone")
}

func TestUnavailableIdentityProviderGives503(t *testing.T) {
	f := newFixtureWithAuth(t, &stubIDP{err: auth.ErrProviderUnavailable})

	status, _ := f.call(t, ola, http.MethodGet, "/v1/me", nil)

	assert.Equal(t, http.StatusServiceUnavailable, status)
}

func TestSwaggerSpecIsServed(t *testing.T) {
	f := newFixture(t)

	resp, err := http.Get(f.server.URL + "/swagger/doc.json") //nolint:noctx // test
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&spec))
	assert.Contains(t, spec.Paths, "/me/schedule")
	assert.Contains(t, spec.Paths, "/admin/red-days/{date}")
}
