-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE personnel (
    id       BIGSERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    email    TEXT NOT NULL,
    groups   TEXT[] NOT NULL DEFAULT '{}',
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE UNIQUE INDEX personnel_email_key ON personnel (lower(email));

-- group_name is the OIDC group whose members belong to the team. manager_id is the person in charge of the team and its shifts.
CREATE TABLE team (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    group_name TEXT UNIQUE,
    manager_id BIGINT REFERENCES personnel (id) ON DELETE SET NULL
);

-- Everyone in a team may man and cover the team's positions. Rows of teams with a group_name are synced from personnel.groups.
CREATE TABLE team_member (
    team_id   BIGINT NOT NULL REFERENCES team (id) ON DELETE CASCADE,
    person_id BIGINT NOT NULL REFERENCES personnel (id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, person_id)
);
CREATE INDEX team_member_person_idx ON team_member (person_id);

-- A position is a named slot in a team, manned by one person at a time.
CREATE TABLE position (
    id      BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES team (id) ON DELETE CASCADE,
    name    TEXT NOT NULL,
    kind    TEXT NOT NULL,
    UNIQUE (team_id, name)
);

-- Who holds a position between starts_at and ends_at. A missing assignee_id means the period is not filled yet.
-- The backend currently creates one-week instances (Monday 08:00 to Monday 08:00), but nothing here requires it.
CREATE TABLE assignment (
    id          BIGSERIAL PRIMARY KEY,
    position_id BIGINT NOT NULL REFERENCES position (id) ON DELETE CASCADE,
    starts_at   TIMESTAMPTZ NOT NULL,
    ends_at     TIMESTAMPTZ NOT NULL,
    assignee_id BIGINT REFERENCES personnel (id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  BIGINT REFERENCES personnel (id) ON DELETE SET NULL,
    CHECK (ends_at > starts_at),
    UNIQUE (position_id, starts_at),
    EXCLUDE USING gist (position_id WITH =, tstzrange(starts_at, ends_at) WITH &&)
);
CREATE INDEX assignment_assignee_idx ON assignment (assignee_id, starts_at);

-- Someone temporarily taking over an assignment, for the rest of a day or the whole instance.
CREATE TABLE cover (
    id            BIGSERIAL PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES assignment (id) ON DELETE CASCADE,
    covered_by    BIGINT NOT NULL REFERENCES personnel (id),
    starts_at     TIMESTAMPTZ NOT NULL,
    ends_at       TIMESTAMPTZ NOT NULL,
    reason        TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by    BIGINT REFERENCES personnel (id) ON DELETE SET NULL,
    CHECK (ends_at > starts_at),
    EXCLUDE USING gist (assignment_id WITH =, tstzrange(starts_at, ends_at) WITH &&)
);
CREATE INDEX cover_covered_by_idx ON cover (covered_by, ends_at);

-- Admin overrides of the computed red days (weekends and Norwegian public holidays).
CREATE TABLE red_day_override (
    day        DATE PRIMARY KEY,
    name       TEXT NOT NULL DEFAULT '',
    is_red_day BOOLEAN NOT NULL
);

-- +goose Down
DROP TABLE red_day_override;
DROP TABLE cover;
DROP TABLE assignment;
DROP TABLE position;
DROP TABLE team_member;
DROP TABLE team;
DROP TABLE personnel;
