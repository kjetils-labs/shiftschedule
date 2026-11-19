package devdata_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/shiftschedule/test/devdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestTeamsAreUsableForCovers(t *testing.T) {
	emails := map[string]bool{}
	for _, p := range devdata.People {
		assert.False(t, emails[p.Email], "duplicate email %s", p.Email)
		emails[p.Email] = true
	}
	assert.True(t, emails[devdata.AdminEmail])

	for _, team := range devdata.Teams {
		assert.True(t, emails[team.ManagerEmail], "manager of %s must be a person", team.Name)

		members := 0
		for _, p := range devdata.People {
			if slices.Contains(p.Groups, team.Group) {
				members++
			}
		}
		assert.GreaterOrEqual(t, members, 2, "%s needs two members, or nobody can cover for anybody", team.Name)
	}
}

func TestDexConfigMatchesTheCommittedFile(t *testing.T) {
	want, err := devdata.DexConfig()
	require.NoError(t, err)

	got, err := os.ReadFile("../../hacks/dex/config.yaml")
	require.NoError(t, err)

	assert.Equal(t, string(want), string(got), "hacks/dex/config.yaml is out of date, run: task dex:config")
}

func TestPasswordHashMatchesPassword(t *testing.T) {
	config, err := devdata.DexConfig()
	require.NoError(t, err)

	var hash string
	for line := range strings.SplitSeq(string(config), "\n") {
		if rest, ok := strings.CutPrefix(line, `    hash: "`); ok {
			hash = strings.TrimSuffix(rest, `"`)
			break
		}
	}
	require.NotEmpty(t, hash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte(devdata.Password)))
}
