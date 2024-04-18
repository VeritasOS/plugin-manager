package pluginmanager

// Status of plugin execution used for displaying to user on console.
const (
	DStatusFail  = "Failed"
	DStatusOk    = "Succeeded"
	DStatusSkip  = "Skipped"
	DStatusStart = "Starting"
)

// PluginAttributes that are supported in a plugin file.
type PluginAttributes struct {
	Description string
	FileName    string
	ExecStart   string
	RequiredBy  []string
	Requires    []string
}

// Plugins is basically a map of file and its contents.
type Plugins map[string]*PluginAttributes

// RunStatus is the plugin run's info: status, stdouterr.
type RunStatus struct {
	PluginAttributes `yaml:",inline"`
	Status           string
	StdOutErr        string
}

// RunAllStatus is the pm run status.
type RunAllStatus struct {
	Type string
	// TODO: Add Percentage to get no. of pending vs. completed run of plugins.
	Plugins   PluginsStatus `yaml:",omitempty"`
	Status    string
	StdOutErr string
}

// PluginsStatus is a list of plugins' run info.
type PluginsStatus []RunStatus

// Workflow indicates action and rollback plugin types to be run.
type Workflow []ActionRollback

// ActionRollback basically contains plugin-type info for Action and Rollback.
type ActionRollback struct {
	Action   string
	Rollback string `yaml:",omitempty"`
}

// WorkflowStatus contains status info for Workflow.
type WorkflowStatus struct {
	Status    string
	StdOutErr string
	// TODO: Add Percentage to get no. of pending vs. completed run of plugins.
	// ActionRollbacks []ActionRollbackStatus `yaml:",omitempty"`
	Action   []RunAllStatus `yaml:",omitempty"`
	Rollback []RunAllStatus `yaml:",omitempty"`
}

// ActionRollbackStatus contains run status of Action, and contains Rollback plugins' status if specified in case of Action's failure.
type ActionRollbackStatus struct {
	Action   RunAllStatus `yaml:",omitempty"`
	Rollback RunAllStatus `yaml:",omitempty"`
}
