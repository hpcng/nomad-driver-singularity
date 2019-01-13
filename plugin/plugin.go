// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package plugin

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	firecracker "github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/google/uuid"
	hclog "github.com/hashicorp/go-hclog"
	cstructs "github.com/hashicorp/nomad/client/structs"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/hashicorp/nomad/plugins/shared/loader"
	pstructs "github.com/hashicorp/nomad/plugins/shared/structs"
)

const (
	// pluginName is the name of the plugin
	pluginName = "singularity"

	// fingerprintPeriod is the interval at which the driver will send fingerprint responses
	fingerprintPeriod = 30 * time.Second

	// singularityVersion is the earliest supported version of singularity
	singularityVersion = "v3.0.0"

	// singularityCmd is the command singularity is installed as.
	singularityCmd = "singularity"
)

var (
	// PluginID is the exec plugin metadata registered in the plugin
	// catalog.
	PluginID = loader.PluginID{
		Name:       pluginName,
		PluginType: base.PluginTypeDriver,
	}

	// PluginConfig is the exec driver factory function registered in the
	// plugin catalog.
	PluginConfig = &loader.InternalPluginConfig{
		Config:  map[string]interface{}{},
		Factory: func(l hclog.Logger) interface{} { return NewDriver(l) },
	}

	// pluginInfo is the response returned for the PluginInfo RPC
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{"0.1.0"},
		PluginVersion:     "0.0.1",
		Name:              pluginName,
	}

	// configSpec is the hcl specification returned by the ConfigSchema RPC
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"volumes_enabled": hclspec.NewDefault(
			hclspec.NewAttr("volumes_enabled", "bool", false),
			hclspec.NewLiteral("true"),
		),
	})

	// taskConfigSpec is the hcl specification for the driver config section of
	// a taskConfig within a job. It is returned in the TaskConfigSchema RPC
	taskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"image":   hclspec.NewAttr("image", "string", true),
		"command": hclspec.NewAttr("command", "string", false),
		"args":    hclspec.NewAttr("args", "list(string)", false),

		"binds":   hclspec.NewAttr("bind", "list(string)", false),
		"contain": hclspec.NewAttr("contain", "bool", false),
		"home":    hclspec.NewAttr("home", "string", false),
		"workdir": hclspec.NewAttr("workdir", "string", false),
		"pwd":     hclspec.NewAttr("pwd", "string", false),
		"debug":   hclspec.NewAttr("debug", "bool", false),
	})

	// capabilities is returned by the Capabilities RPC and indicates what
	// optional features this driver supports
	capabilities = &drivers.Capabilities{
		SendSignals: true,
		Exec:        false,
		FSIsolation: cstructs.FSIsolationChroot,
	}
)

type DriverConfig struct {
	// VolumesEnabled allows tasks to bind host paths (volumes) inside their
	// container. Binding relative paths is always allowed and will be resolved
	// relative to the allocation's directory.
	VolumesEnabled bool `codec:"volumes_enabled"`
}

type TaskConfig struct {
	ImageName string   `codec:"image"`
	Command   string   `codec:"command"`
	Args      []string `codec:"args"`

	Binds   []string `codec:"binds"` // Host-Volumes to mount in, syntax: /path/to/host/directory:/destination/path/in/container
	Contain bool     `codec:"contain"`
	Home    string   `codec:"home"`
	Workdir string   `codec:"workdir"`
	Pwd     string   `codec:"pwd"`
	Debug   bool     `codec:"debug"` // Enable debug option for singularity command
}

// Driver is a driver for running images via singularity
// We attempt to chose sane defaults for now, with more configuration available
// planned in the future
type Driver struct {
	// eventer is used to handle multiplexing of TaskEvents calls such that an
	// event can be broadcast to all callers
	eventer *eventer.Eventer

	// config is the driver configuration set by the SetConfig RPC
	config *Config

	// tasks is the in memory datastore mapping taskIDs to singularityTaskHandles
	tasks *taskStore

	// ctx is the context for the driver. It is passed to other subsystems to
	// coordinate shutdown
	ctx context.Context

	// signalShutdown is called when the driver is shutting down and cancels the
	// ctx passed to any subsystems
	signalShutdown context.CancelFunc

	// logger will log to the plugin output which is usually an 'executor.out'
	// file located in the root of the TaskDir
	logger hclog.Logger
}

var _ drivers.DriverPlugin = &Driver{}

func NewDriver(logger hclog.Logger) *Driver {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)
	return &Driver{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &Config{},
		tasks:          newTaskStore(),
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}

func (*Driver) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

func (*Driver) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

func (d *Driver) SetConfig(cfg *base.Config) error {
	var config DriverConfig
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	d.logger.Warn("Set Config", "config", fmt.Sprintf("%#v", config), "raw", cfg.PluginConfig)

	d.config = &config

	return nil
}

func (d *Driver) Shutdown(ctx context.Context) error {
	return nil
}

func (d *Driver) TaskConfigSchema() (*hclspec.Spec, error) {
	return taskConfigSpec, nil
}

func (d *Driver) Capabilities() (*drivers.Capabilities, error) {
	return capabilities, nil
}

func (d *Driver) Fingerprint(ctx context.Context) (<-chan *drivers.Fingerprint, error) {
	ch := make(chan *drivers.Fingerprint, 1)
	ch <- d.buildFingerprint()
	go d.handleFingerprint(ctx, ch)
	return ch, nil
}

func (d *Driver) handleFingerprint(ctx context.Context, ch chan<- *drivers.Fingerprint) {
	defer close(ch)
	ticker := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ch <- d.buildFingerprint()
			ticker.Reset(fingerprintPeriod)
		}
	}
}

func (d *Driver) fingerprintBinary(path string) *drivers.Fingerprint {
	finfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUndetected,
			HealthDescription: fmt.Sprintf("Binary, %q, does not exist: %v", path, err),
		}
	}

	if err != nil {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUnhealthy,
			HealthDescription: fmt.Sprintf("Failed to stat binary, %q: %v", path, err),
		}
	}

	if finfo.IsDir() {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUndetected,
			HealthDescription: fmt.Sprintf("Binary, %q is a directory", path),
		}
	} else if finfo.Mode()&executableMask == 0 {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUnhealthy,
			HealthDescription: fmt.Sprintf("Binary, %q, is not executable. Check permissions of binary", path),
		}
	}

	return nil
}

func (d *Driver) buildFingerprint() *drivers.Fingerprint {
	if d.config == nil {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUnhealthy,
			HealthDescription: "Waiting for config",
		}
	}

	if f := d.fingerprintBinary(d.config.FirecrackerPath); f != nil {
		return f
	}

	if d.config.UseJailer == true {
		if f := d.fingerprintBinary(d.config.JailerPath); f != nil {
			return f
		}
	}

	health := drivers.HealthStateHealthy
	desc := "ready"
	attrs := map[string]*pstructs.Attribute{"driver.firecracker": pstructs.NewStringAttribute("1")}

	return &drivers.Fingerprint{
		Attributes:        attrs,
		Health:            health,
		HealthDescription: desc,
	}
}

func (d *Driver) RecoverTask(handle *drivers.TaskHandle) error {
	return nil
}

func newFirecracker(ctx context.Context, binPath, socketPath, kernelImage, kernelArgs, fsPath string, cpuCount int64, memSize int64) (*firecracker.Machine, error) {
	rootDrive := models.Drive{
		DriveID:      firecracker.String("1"),
		PathOnHost:   firecracker.String(fsPath),
		IsRootDevice: firecracker.Bool(true),
		IsReadOnly:   firecracker.Bool(true),
	}

	fcCfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelImage,
		KernelArgs:      kernelArgs,
		Drives:          []models.Drive{rootDrive},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:   cpuCount,
			CPUTemplate: models.CPUTemplate("C3"),
			HtEnabled:   false,
			MemSizeMib:  memSize,
		},
	}

	machineOpts := []firecracker.Opt{}

	// TODO: Support jailer
	cmd := firecracker.VMCommandBuilder{}.
		WithBin(binPath).
		WithSocketPath(socketPath).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		Build(ctx)
	machineOpts = append(machineOpts, firecracker.WithProcessRunner(cmd))

	m, err := firecracker.NewMachine(ctx, fcCfg, machineOpts...)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (d *Driver) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *cstructs.DriverNetwork, error) {
	if _, ok := d.tasks.Get(cfg.ID); ok {
		return nil, nil, fmt.Errorf("task with ID %q already started", cfg.ID)
	}

	ctx := context.Background()
	handle := drivers.NewTaskHandle(pluginName)

	var config TaskConfig
	if err := cfg.DecodeDriverConfig(&config); err != nil {
		return nil, nil, err
	}

	if config.KernelBootArgs == "" {
		config.KernelBootArgs = defaultBootArgs
	}

	controlUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, nil, err
	}

	cpuCount := int64(math.Max(1, float64(cfg.Resources.NomadResources.Cpu.CpuShares)/1024.0))
	memSize := cfg.Resources.NomadResources.Memory.MemoryMB

	controlSocketPath := fmt.Sprintf("/tmp/%s.socket", controlUUID)
	m, err := newFirecracker(ctx, d.config.FirecrackerPath, controlSocketPath, config.KernelPath, config.KernelBootArgs, config.ImagePath, cpuCount, memSize)
	if err != nil {
		return nil, nil, err
	}

	h := &taskHandle{
		taskConfig: cfg,
		machine:    m,
		procState:  drivers.TaskStateRunning,
		startedAt:  time.Now().Round(time.Millisecond),
		logger:     d.logger,
		waitCh:     make(chan struct{}),
	}

	d.tasks.Set(cfg.ID, h)
	go h.run()
	return handle, nil, nil
}

func (d *Driver) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	h, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, fmt.Errorf("task with ID %q not found", taskID)
	}

	ch := make(chan *drivers.ExitResult)
	go func(ch chan *drivers.ExitResult, task *taskHandle) {
		<-task.waitCh
		ch <- task.exitResult
	}(ch, h)

	return ch, nil
}

func (d *Driver) StopTask(taskID string, timeout time.Duration, signal string) error {
	h, ok := d.tasks.Get(taskID)
	if !ok {
		return fmt.Errorf("task with ID %q not found", taskID)
	}

	return h.machine.StopVMM()
}

func (d *Driver) DestroyTask(taskID string, force bool) error {
	d.tasks.Delete(taskID)

	// TODO: Destroy any ephemeral storage and ensure firecracker proc is dead.
	return nil
}

func (d *Driver) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	h, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, fmt.Errorf("task with ID %q not found", taskID)
	}

	return h.TaskStatus(), nil
}

func (d *Driver) TaskStats(taskID string) (*cstructs.TaskResourceUsage, error) {
	_, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, fmt.Errorf("task with ID %q not found", taskID)
	}

	return &cstructs.TaskResourceUsage{
		ResourceUsage: &cstructs.ResourceUsage{
			MemoryStats: &cstructs.MemoryStats{},
			CpuStats:    &cstructs.CpuStats{},
		},
		Pids: make(map[string]*cstructs.ResourceUsage),
	}, nil
}

func (d *Driver) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	return make(chan *drivers.TaskEvent), nil
}

func (d *Driver) SignalTask(taskID string, signal string) error {
	return nil
}

func (d *Driver) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	return nil, nil
}
