package planner

import "time"

type Pipeline struct {
	Metadata Metadata `json:"metadata"`
	Spec     struct {
		Tasks      []PipelineTask `json:"tasks"`
		Workspaces []Workspace    `json:"workspaces"`
	} `json:"spec"`
}

type Workspace struct {
	Name     string `json:"name"`
	ReadOnly bool   `json:"readOnly,omitempty"`
}

type Metadata struct {
	Name              string            `json:"name"`
	UID               string            `json:"uid"`
	Generation        int64             `json:"generation"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels"`
}

type PipelineTask struct {
	Name       string             `json:"name"`
	RunAfter   []string           `json:"runAfter,omitempty"`
	Params     []Param            `json:"params,omitempty"`
	Workspaces []WorkspaceBinding `json:"workspaces,omitempty"`
	TaskRef    map[string]any     `json:"taskRef,omitempty"`
	TaskSpec   map[string]any     `json:"taskSpec,omitempty"`
	When       []map[string]any   `json:"when,omitempty"`
}

type Param struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type WorkspaceBinding struct {
	Name      string `json:"name"`
	Workspace string `json:"workspace"`
}

type PipelineRun struct {
	Metadata Metadata `json:"metadata"`
	Spec     struct {
		PipelineRef struct {
			Name string `json:"name"`
		} `json:"pipelineRef"`
		Workspaces []map[string]any `json:"workspaces,omitempty"`
	} `json:"spec"`
	Status struct {
		PipelineSpec *struct {
			Tasks      []PipelineTask `json:"tasks"`
			Workspaces []Workspace    `json:"workspaces"`
		} `json:"pipelineSpec"`
		ChildReferences []ChildReference `json:"childReferences"`
	} `json:"status"`
}

type ChildReference struct {
	Name         string `json:"name"`
	PipelineTask string `json:"pipelineTaskName"`
	Status       string `json:"status"`
}

type TaskRunList struct {
	Items []TaskRun `json:"items"`
}

type TaskRun struct {
	Metadata Metadata `json:"metadata"`
	Status   struct {
		Conditions []Condition  `json:"conditions"`
		Results    []TaskResult `json:"results"`
	} `json:"status"`
}

type Condition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type TaskResult struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type Policy struct {
	Metadata Metadata `json:"metadata"`
	Spec     struct {
		PipelineRef     string              `json:"pipelineRef"`
		ResumeWithin    string              `json:"resumeWithin"`
		RetryOnlyLatest bool                `json:"retryOnlyLatest"`
		Tasks           map[string]TaskRule `json:"tasks"`
	} `json:"spec"`
}

type TaskRule struct {
	RetryWith    []string `json:"retryWith"`
	BlocksResume bool     `json:"blocksResume"`
}

type Plan struct {
	SourceRun        string     `json:"sourceRun"`
	SourceRunUID     string     `json:"sourceRunUID,omitempty"`
	Pipeline         string     `json:"pipeline"`
	Policy           string     `json:"policy"`
	PolicyGeneration int64      `json:"policyGeneration,omitempty"`
	Decision         string     `json:"decision"`
	Tasks            []TaskPlan `json:"tasks"`
	Warnings         []Warning  `json:"warnings,omitempty"`
	RefusalReason    string     `json:"refusalReason,omitempty"`
	Metrics          Metrics    `json:"metrics"`
}

type Warning struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Tasks     []string `json:"tasks,omitempty"`
	StateKind string   `json:"stateKind,omitempty"`
	StateName string   `json:"stateName,omitempty"`
}

type Metrics struct {
	TotalTasks       int     `json:"totalTasks"`
	RerunTasks       int     `json:"rerunTasks"`
	InheritedTasks   int     `json:"inheritedTasks"`
	ContinuingTasks  int     `json:"continuingTasks"`
	TasksAvoided     int     `json:"tasksAvoided"`
	AvoidancePercent float64 `json:"avoidancePercent"`
}

type TaskPlan struct {
	Name          string `json:"name"`
	Action        string `json:"action"`
	Reason        string `json:"reason"`
	SourceTaskRun string `json:"sourceTaskRun,omitempty"`
}
