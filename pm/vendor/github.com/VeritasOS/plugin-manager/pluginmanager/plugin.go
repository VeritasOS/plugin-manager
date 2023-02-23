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
