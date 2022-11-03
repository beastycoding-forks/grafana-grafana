package pluginutils

import (
	"testing"

	"github.com/grafana/grafana/pkg/plugins"
	ac "github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/stretchr/testify/require"
)

func TestToRegistrations(t *testing.T) {
	tests := []struct {
		name string
		regs []plugins.RoleRegistration
		want []ac.RoleRegistration
	}{
		{
			name: "no registration",
			regs: nil,
			want: []ac.RoleRegistration{},
		},
		{
			name: "registration gets converted successfully",
			regs: []plugins.RoleRegistration{
				{
					Role: plugins.Role{
						Name:        "Tester",
						Description: "Test",
						Permissions: []plugins.Permission{
							{Action: "test:action"},
							{Action: "test:action", Scope: "test:scope"},
						},
					},
					Grants: []string{"Admin", "Editor"},
				},
				{
					Role: plugins.Role{
						Name:        "Admin Validator",
						Permissions: []plugins.Permission{},
					},
				},
			},
			want: []ac.RoleRegistration{
				{
					Role: ac.RoleDTO{
						Version:     1,
						Name:        ac.AppPluginRolePrefix + "plugin-id:tester",
						DisplayName: "Tester",
						Description: "Test",
						Group:       "Plugin Name",
						Permissions: []ac.Permission{
							{Action: "test:action"},
							{Action: "test:action", Scope: "test:scope"},
						},
						OrgID: ac.GlobalOrgID,
					},
					Grants: []string{"Admin", "Editor"},
				},
				{
					Role: ac.RoleDTO{
						Version:     1,
						Name:        ac.AppPluginRolePrefix + "plugin-id:admin-validator",
						DisplayName: "Admin Validator",
						Group:       "Plugin Name",
						Permissions: []ac.Permission{},
						OrgID:       ac.GlobalOrgID,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToRegistrations("plugin-id", "Plugin Name", tt.regs)
			require.Equal(t, tt.want, got)
		})
	}
}
