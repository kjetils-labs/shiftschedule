// Package devdata is the one description of the people, teams and positions used in development.
// The generator seeds the database from it and renders the Dex config from it, so logins and personnel always match.
package devdata

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // stable ids, not security
	_ "embed"     // dexTemplate
	"fmt"
	"strings"
	"text/template"

	"github.com/shiftschedule/internal/domain"
)

// Password is the password of every development user in Dex.
const Password = "password"

// passwordHash is the bcrypt hash of Password.
const passwordHash = "$2a$10$2umYE.3BOctrgPv7x/nD1e8idYJeqKu2xaE.UPcdlf3LbIFSa6k5C"

// Groups, as the identity provider reports them. A person is in a team when they have the team's group.
const (
	GroupServer   = "server@domain.no"
	GroupNetwork  = "network@domain.no"
	GroupDatabase = "database@domain.no"
	GroupKlient   = "klient@domain.no"
)

const AdminEmail = "admin@example.no"

// Person is a development user. The email is both the Dex login and the key in personnel.
type Person struct {
	Name    string
	Email   string
	Groups  []string
	IsAdmin bool
}

// Position is a slot in a team.
type Position struct {
	Name string
	Kind domain.ShiftKind
}

// Team is a group of people and the positions they man. ManagerEmail is the email of a Person.
type Team struct {
	Name         string
	Group        string
	ManagerEmail string
	Positions    []Position
}

// People is ordered, so everything derived from it is deterministic. Every team has at least two members, because covers stay inside a team.
var People = []Person{
	{Name: "Kari Admin", Email: AdminEmail, IsAdmin: true},
	{Name: "Kari Nordmann", Email: "kari@example.no", Groups: []string{GroupServer, GroupDatabase}},
	{Name: "Ola Nordmann", Email: "ola@example.no", Groups: []string{GroupServer, GroupNetwork, GroupKlient}},
	{Name: "Per Hansen", Email: "per@example.no", Groups: []string{GroupNetwork, GroupDatabase, GroupKlient}},
}

// Teams are managed by Kari Nordmann.
var Teams = []Team{
	{Name: "Team Server", Group: GroupServer, ManagerEmail: "kari@example.no", Positions: dayAndOnCall()},
	{Name: "Team Network", Group: GroupNetwork, ManagerEmail: "kari@example.no", Positions: dayAndOnCall()},
	{Name: "Team Database", Group: GroupDatabase, ManagerEmail: "kari@example.no", Positions: dayAndOnCall()},
	{Name: "Team Klient", Group: GroupKlient, ManagerEmail: "kari@example.no", Positions: []Position{{"Dagvakt", domain.KindDayShift}}},
}

func dayAndOnCall() []Position {
	return []Position{{"Dagvakt", domain.KindDayShift}, {"Vakt", domain.KindOnCall}}
}

// PositionKey identifies a position by team and name, e.g. "Team Server/Vakt".
func PositionKey(team, position string) string {
	return team + "/" + position
}

//go:embed dex.yaml.tmpl
var dexTemplate string

// DexConfig renders the Dex config for local development, with a static user for every Person.
func DexConfig() ([]byte, error) {
	type user struct{ Email, Username, UserID, Hash string }
	users := make([]user, 0, len(People))
	for _, p := range People {
		username, _, _ := strings.Cut(p.Email, "@")
		users = append(users, user{Email: p.Email, Username: username, UserID: stableID(p.Email), Hash: passwordHash})
	}

	tmpl, err := template.New("dex").Parse(dexTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the Dex template: %w", err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, map[string]any{"Users": users}); err != nil {
		return nil, fmt.Errorf("failed to render the Dex config: %w", err)
	}
	return out.Bytes(), nil
}

// stableID is a name-based UUID (version 5 layout) so a user keeps the same id every time the config is rendered.
func stableID(email string) string {
	sum := sha1.Sum([]byte("shiftschedule-dev/" + email)) //nolint:gosec // stable ids, not security
	sum[6] = sum[6]&0x0f | 0x50
	sum[8] = sum[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}
