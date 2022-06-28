// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux
// +build linux

//go:generate go run github.com/tinylib/msgp -o=activity_dump_gen_linux.go -tests=false
//go:generate go run github.com/mailru/easyjson/easyjson -gen_build_flags=-mod=mod -no_std_marshalers -build_tags linux $GOFILE

package probe

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/DataDog/gopsutil/process"
	"github.com/prometheus/procfs"
	"github.com/tinylib/msgp/msgp"
	"go.uber.org/atomic"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/DataDog/datadog-agent/pkg/process/util"
	adproto "github.com/DataDog/datadog-agent/pkg/security/adproto/v1"
	"github.com/DataDog/datadog-agent/pkg/security/api"
	"github.com/DataDog/datadog-agent/pkg/security/ebpf/probes"
	seclog "github.com/DataDog/datadog-agent/pkg/security/log"
	"github.com/DataDog/datadog-agent/pkg/security/metrics"
	"github.com/DataDog/datadog-agent/pkg/security/probe/dump"
	"github.com/DataDog/datadog-agent/pkg/security/secl/compiler/eval"
	"github.com/DataDog/datadog-agent/pkg/security/secl/model"
	"github.com/DataDog/datadog-agent/pkg/security/utils"
	"github.com/DataDog/datadog-agent/pkg/version"
)

const (
	// ProtobufVersion defines the protobuf version in use
	ProtobufVersion = "v1"
	// ActivityDumpSource defines the source of activity dumps
	ActivityDumpSource = "runtime-security-agent"
)

// ActivityDumpStatus defines the state of an activity dump
type ActivityDumpStatus int

const (
	// Stopped means that the ActivityDump is not active
	Stopped ActivityDumpStatus = iota
	// Running means that the ActivityDump is active
	Running
)

// DumpMetadata is used to provide context about the activity dump
type DumpMetadata struct {
	AgentVersion      string `msg:"agent_version" json:"agent_version"`
	AgentCommit       string `msg:"agent_commit" json:"agent_commit"`
	KernelVersion     string `msg:"kernel_version" json:"kernel_version"`
	LinuxDistribution string `msg:"linux_distribution" json:"linux_distribution"`
	Arch              string `msg:"arch" json:"arch"`

	Name              string        `msg:"name" json:"name"`
	ProtobufVersion   string        `msg:"protobuf_version" json:"protobuf_version"`
	DifferentiateArgs bool          `msg:"differentiate_args" json:"differentiate_args"`
	Comm              string        `msg:"comm,omitempty" json:"comm,omitempty"`
	ContainerID       string        `msg:"container_id,omitempty" json:"-"`
	Start             time.Time     `msg:"start" json:"start"`
	Timeout           time.Duration `msg:"-" json:"-"`
	End               time.Time     `msg:"end" json:"end"`
	timeoutRaw        int64         `msg:"-"`
	Size              uint64        `msg:"activity_dump_size,omitempty" json:"activity_dump_size,omitempty"`
	Serialization     string        `msg:"serialization,omitempty" json:"serialization,omitempty"`
}

// ActivityDump holds the activity tree for the workload defined by the provided list of tags. The encoding described by
// the `msg` annotation is used to generate the activity dump file while the encoding described by the `json` annotation
// is used to generate the activity dump metadata sent to the event platform.
// easyjson:json
type ActivityDump struct {
	*sync.Mutex        `msg:"-"`
	state              ActivityDumpStatus
	adm                *ActivityDumpManager
	processedCount     map[model.EventType]*atomic.Uint64
	addedRuntimeCount  map[model.EventType]*atomic.Uint64
	addedSnapshotCount map[model.EventType]*atomic.Uint64

	// standard attributes used by the intake
	Host    string   `msg:"host" json:"host,omitempty"`
	Service string   `msg:"service,omitempty" json:"service,omitempty"`
	Source  string   `msg:"source" json:"ddsource,omitempty"`
	Tags    []string `msg:"tags,omitempty" json:"-"`
	DDTags  string   `msg:"-" json:"ddtags,omitempty"`

	CookiesNode         map[uint32]*ProcessActivityNode              `msg:"-" json:"-"`
	ProcessActivityTree []*ProcessActivityNode                       `msg:"tree,omitempty" json:"-"`
	StorageRequests     map[dump.StorageFormat][]dump.StorageRequest `msg:"storage_requests,omitempty" json:"-"`

	// Dump metadata
	DumpMetadata `msg:"metadata"`
}

// NewEmptyActivityDump returns a new zero-like instance of an ActivityDump
func NewEmptyActivityDump() *ActivityDump {
	return &ActivityDump{
		Mutex: &sync.Mutex{},
	}
}

// WithDumpOption can be used to configure an ActivityDump
//msgp:ignore WithDumpOption
type WithDumpOption func(ad *ActivityDump)

// NewActivityDump returns a new instance of an ActivityDump
func NewActivityDump(adm *ActivityDumpManager, options ...WithDumpOption) *ActivityDump {
	ad := ActivityDump{
		Mutex: &sync.Mutex{},
		DumpMetadata: DumpMetadata{
			AgentVersion:      version.AgentVersion,
			AgentCommit:       version.Commit,
			KernelVersion:     adm.probe.kernelVersion.Code.String(),
			LinuxDistribution: adm.probe.kernelVersion.OsRelease["PRETTY_NAME"],
			Name:              fmt.Sprintf("activity-dump-%s", eval.RandString(10)),
			ProtobufVersion:   ProtobufVersion,
			Start:             time.Now(),
			Arch:              probes.RuntimeArch,
		},
		Host:               adm.hostname,
		Source:             ActivityDumpSource,
		CookiesNode:        make(map[uint32]*ProcessActivityNode),
		adm:                adm,
		processedCount:     make(map[model.EventType]*atomic.Uint64),
		addedRuntimeCount:  make(map[model.EventType]*atomic.Uint64),
		addedSnapshotCount: make(map[model.EventType]*atomic.Uint64),
		StorageRequests:    make(map[dump.StorageFormat][]dump.StorageRequest),
	}

	for _, option := range options {
		option(&ad)
	}

	// generate counters
	for i := model.EventType(0); i < model.MaxKernelEventType; i++ {
		ad.processedCount[i] = atomic.NewUint64(0)
		ad.addedRuntimeCount[i] = atomic.NewUint64(0)
		ad.addedSnapshotCount[i] = atomic.NewUint64(0)
	}
	return &ad
}

// NewActivityDumpFromMessage returns a new ActivityDump from a SecurityActivityDumpMessage
func NewActivityDumpFromMessage(msg *api.ActivityDumpMessage) (*ActivityDump, error) {
	metadata := msg.GetMetadata()
	if metadata == nil {
		return nil, fmt.Errorf("couldn't create new ActivityDump: missing activity dump metadata")
	}

	startTime, err := time.Parse(time.RFC822, metadata.GetStart())
	if err != nil {
		return nil, fmt.Errorf("couldn't parse start time [%s]: %w", metadata.GetStart(), err)
	}
	timeout, err := time.ParseDuration(metadata.GetTimeout())
	if err != nil {
		return nil, fmt.Errorf("couldn't parse timeout [%s]: %w", metadata.GetTimeout(), err)
	}

	ad := ActivityDump{
		Mutex:              &sync.Mutex{},
		CookiesNode:        make(map[uint32]*ProcessActivityNode),
		processedCount:     make(map[model.EventType]*atomic.Uint64),
		addedRuntimeCount:  make(map[model.EventType]*atomic.Uint64),
		addedSnapshotCount: make(map[model.EventType]*atomic.Uint64),
		StorageRequests:    make(map[dump.StorageFormat][]dump.StorageRequest),
		Host:               msg.GetHost(),
		Service:            msg.GetService(),
		Source:             msg.GetSource(),
		Tags:               msg.GetTags(),
		DumpMetadata: DumpMetadata{
			AgentVersion:      metadata.GetAgentVersion(),
			AgentCommit:       metadata.GetAgentCommit(),
			KernelVersion:     metadata.GetKernelVersion(),
			LinuxDistribution: metadata.GetLinuxDistribution(),
			Name:              metadata.GetName(),
			ProtobufVersion:   metadata.GetProtobufVersion(),
			DifferentiateArgs: metadata.GetDifferentiateArgs(),
			Comm:              metadata.GetComm(),
			ContainerID:       metadata.GetContainerID(),
			Start:             startTime,
			Timeout:           timeout,
			End:               startTime.Add(timeout),
			Size:              metadata.GetSize(),
			Arch:              metadata.GetArch(),
		},
	}

	// parse requests from message
	for _, request := range msg.GetStorage() {
		storageType, err := dump.ParseStorageType(request.GetType())
		if err != nil {
			// invalid storage type, ignore
			continue
		}
		storageFormat, err := dump.ParseStorageFormat(request.GetFormat())
		if err != nil {
			// invalid storage format, ignore
			continue
		}
		ad.StorageRequests[storageFormat] = append(ad.StorageRequests[storageFormat], dump.NewStorageRequest(
			storageType,
			storageFormat,
			request.GetCompression(),
			filepath.Base(request.File),
		))
	}
	return &ad, nil
}

// SetState sets the status of the activity dump
func (ad *ActivityDump) SetState(state ActivityDumpStatus) {
	ad.Lock()
	defer ad.Unlock()
	ad.state = state
}

// AddStorageRequest adds a storage request to an activity dump
func (ad *ActivityDump) AddStorageRequest(request dump.StorageRequest) {
	ad.Lock()
	defer ad.Unlock()

	if ad.StorageRequests == nil {
		ad.StorageRequests = make(map[dump.StorageFormat][]dump.StorageRequest)
	}
	ad.StorageRequests[request.Format] = append(ad.StorageRequests[request.Format], request)
}

// getTimeoutRawTimestamp returns the timeout timestamp of the current activity dump as a monolitic kernel timestamp
func (ad *ActivityDump) getTimeoutRawTimestamp() int64 {
	if ad.DumpMetadata.timeoutRaw == 0 {
		ad.DumpMetadata.timeoutRaw = ad.adm.probe.resolvers.TimeResolver.ComputeMonotonicTimestamp(ad.DumpMetadata.Start.Add(ad.DumpMetadata.Timeout))
	}
	return ad.DumpMetadata.timeoutRaw
}

// updateTracedPidTimeout updates the timeout of a traced pid in kernel space
func (ad *ActivityDump) updateTracedPidTimeout(pid uint32) {
	// start by looking up any existing entry
	var timeout int64
	_ = ad.adm.tracedPIDsMap.Lookup(pid, &timeout)
	if timeout < ad.getTimeoutRawTimestamp() {
		_ = ad.adm.tracedPIDsMap.Put(pid, ad.getTimeoutRawTimestamp())
	}
}

// commMatches returns true if the ActivityDump comm matches the provided comm
func (ad *ActivityDump) commMatches(comm string) bool {
	return ad.DumpMetadata.Comm == comm
}

// containerIDMatches returns true if the ActivityDump container ID matches the provided container ID
func (ad *ActivityDump) containerIDMatches(containerID string) bool {
	return ad.DumpMetadata.ContainerID == containerID
}

// Matches returns true if the provided list of tags and / or the provided comm match the current ActivityDump
func (ad *ActivityDump) Matches(entry *model.ProcessCacheEntry) bool {
	if entry == nil {
		return false
	}

	if len(ad.DumpMetadata.ContainerID) > 0 {
		if !ad.containerIDMatches(entry.ContainerID) {
			return false
		}
	}

	if len(ad.DumpMetadata.Comm) > 0 {
		if !ad.commMatches(entry.Comm) {
			return false
		}
	}

	return true
}

// scrubAndRetainProcessArgsEnvs scrubs process arguments and environment variables
func (ad *ActivityDump) scrubAndRetainProcessArgsEnvs(nodes []*ProcessActivityNode) {
	// iterate through all the process nodes
	for _, node := range nodes {

		// scrub the current process
		node.scrubAndReleaseArgsEnvs(ad.adm.probe.resolvers.ProcessResolver)

		// scrub each child recursively
		ad.scrubAndRetainProcessArgsEnvs(node.Children)
	}
}

// Stop stops an active dump
func (ad *ActivityDump) Stop() {
	ad.Lock()
	defer ad.Unlock()
	ad.state = Stopped
	ad.DumpMetadata.End = time.Now()

	// remove comm from kernel space
	if len(ad.DumpMetadata.Comm) > 0 {
		commB := make([]byte, 16)
		copy(commB, ad.DumpMetadata.Comm)
		err := ad.adm.tracedCommsMap.Delete(commB)
		if err != nil {
			seclog.Debugf("couldn't delete activity dump filter comm(%s): %v", ad.DumpMetadata.Comm, err)
		}
	}

	// remove container ID from kernel space
	if len(ad.DumpMetadata.ContainerID) > 0 {
		containerIDB := make([]byte, model.ContainerIDLen)
		copy(containerIDB, ad.DumpMetadata.ContainerID)
		err := ad.adm.tracedCgroupsMap.Delete(containerIDB)
		if err != nil {
			seclog.Debugf("couldn't delete activity dump filter containerID(%s): %v", ad.DumpMetadata.ContainerID, err)
		}
	}

	// add additionnal tags
	ad.adm.AddContextTags(ad)

	// look for the service tag and set the service of the dump
	ad.Service = utils.GetTagValue("service", ad.Tags)

	// add the container ID in a tag
	if len(ad.ContainerID) > 0 {
		ad.Tags = append(ad.Tags, "container_id:"+ad.ContainerID)
	}

	// scrub processes and retain args envs now
	ad.scrubAndRetainProcessArgsEnvs(ad.ProcessActivityTree)
}

// nolint: unused
func (ad *ActivityDump) debug() {
	for _, root := range ad.ProcessActivityTree {
		root.debug("")
	}
}

func (ad *ActivityDump) isEventTypeTraced(event *Event) bool {
	// syscall monitor related event
	if event.GetEventType() == model.SyscallsEventType && ad.adm.probe.config.ActivityDumpSyscallMonitor {
		return true
	}

	// other events
	var traced bool
	for _, evtType := range ad.adm.probe.config.ActivityDumpTracedEventTypes {
		if evtType == event.GetEventType() {
			traced = true
		}
	}
	return traced
}

// Insert inserts the provided event in the active ActivityDump. This function returns true if a new entry was added,
// false if the event was dropped.
func (ad *ActivityDump) Insert(event *Event) (newEntry bool) {
	ad.Lock()
	defer ad.Unlock()

	if ad.state != Running {
		// this activity dump is not running anymore, ignore event
		return false
	}

	// ignore fork events for now
	if event.GetEventType() == model.ForkEventType {
		return false
	}

	// metrics
	defer func() {
		if newEntry {
			// this doesn't count the exec events which are counted separately
			ad.addedRuntimeCount[event.GetEventType()].Inc()
		}
	}()

	// find the node where the event should be inserted
	node := ad.findOrCreateProcessActivityNode(event.ResolveProcessCacheEntry(), Runtime)
	if node == nil {
		// a process node couldn't be found for the provided event as it doesn't match the ActivityDump query
		return false
	}

	// check if this event type is traced
	if !ad.isEventTypeTraced(event) {
		return false
	}

	// resolve fields
	event.ResolveFields()

	// the count of processed events is the count of events that matched the activity dump selector = the events for
	// which we successfully found a process activity node
	ad.processedCount[event.GetEventType()].Inc()

	// insert the event based on its type
	switch event.GetEventType() {
	case model.FileOpenEventType:
		return node.InsertFileEvent(&event.Open.File, event, Runtime, ad.adm.probe.config.ActivityDumpPathMergeEnabled)
	case model.DNSEventType:
		return node.InsertDNSEvent(&event.DNS)
	case model.BindEventType:
		return node.InsertBindEvent(&event.Bind)
	case model.SyscallsEventType:
		return node.InsertSyscalls(&event.Syscalls)
	}
	return false
}

// findOrCreateProcessActivityNode finds or a create a new process activity node in the activity dump if the entry
// matches the activity dump selector.
func (ad *ActivityDump) findOrCreateProcessActivityNode(entry *model.ProcessCacheEntry, generationType NodeGenerationType) (node *ProcessActivityNode) {
	if entry == nil {
		return node
	}

	// look for a ProcessActivityNode by process cookie
	if entry.Cookie > 0 {
		var found bool
		node, found = ad.CookiesNode[entry.Cookie]
		if found {
			return node
		}
	}

	defer func() {
		// if a node was found, and if the entry has a valid cookie, insert a cookie shortcut
		if entry.Cookie > 0 && node != nil {
			ad.CookiesNode[entry.Cookie] = node
		}
	}()

	// find or create a ProcessActivityNode for the parent of the input ProcessCacheEntry. If the parent is a fork entry,
	// jump immediately to the next ancestor.
	parentNode := ad.findOrCreateProcessActivityNode(entry.GetNextAncestorNoFork(), Snapshot)

	// if parentNode is nil, the parent of the current node is out of tree (either because the parent is null, or it
	// doesn't match the dump tags).
	if parentNode == nil {

		// since the parent of the current entry wasn't inserted, we need to know if the current entry needs to be inserted.
		if !ad.Matches(entry) {
			return node
		}

		// go through the root nodes and check if one of them matches the input ProcessCacheEntry:
		for _, root := range ad.ProcessActivityTree {
			if root.Matches(entry, ad.DumpMetadata.DifferentiateArgs, ad.adm.probe.resolvers) {
				return root
			}
		}
		// if it doesn't, create a new ProcessActivityNode for the input ProcessCacheEntry
		node = NewProcessActivityNode(entry, generationType)
		// insert in the list of root entries
		ad.ProcessActivityTree = append(ad.ProcessActivityTree, node)

	} else {

		// if parentNode wasn't nil, then (at least) the parent is part of the activity dump. This means that we need
		// to add the current entry no matter if it matches the selector or not. Go through the root children of the
		// parent node and check if one of them matches the input ProcessCacheEntry.
		for _, child := range parentNode.Children {
			if child.Matches(entry, ad.DumpMetadata.DifferentiateArgs, ad.adm.probe.resolvers) {
				return child
			}
		}

		// if none of them matched, create a new ProcessActivityNode for the input processCacheEntry
		node = NewProcessActivityNode(entry, generationType)
		// insert in the list of root entries
		parentNode.Children = append(parentNode.Children, node)
	}

	// count new entry
	switch generationType {
	case Runtime:
		ad.addedRuntimeCount[model.ExecEventType].Inc()
	case Snapshot:
		ad.addedSnapshotCount[model.ExecEventType].Inc()
	}

	// set the pid of the input ProcessCacheEntry as traced
	ad.updateTracedPidTimeout(entry.Pid)

	return node
}

// FindFirstMatchingNode return the first matching node of requested comm
func (ad *ActivityDump) FindFirstMatchingNode(comm string) *ProcessActivityNode {
	for _, node := range ad.ProcessActivityTree {
		if node.Process.Comm == comm {
			return node
		}
	}

	return nil
}

// GetSelectorStr returns a string representation of the profile selector
func (ad *ActivityDump) GetSelectorStr() string {
	ad.Lock()
	defer ad.Unlock()

	return ad.getSelectorStr()
}

// getSelectorStr internal, thread-unsafe version of GetSelectorStr
func (ad *ActivityDump) getSelectorStr() string {
	tags := make([]string, 0, len(ad.Tags)+2)
	if len(ad.DumpMetadata.ContainerID) > 0 {
		tags = append(tags, fmt.Sprintf("container_id:%s", ad.DumpMetadata.ContainerID))
	}
	if len(ad.DumpMetadata.Comm) > 0 {
		tags = append(tags, fmt.Sprintf("comm:%s", ad.DumpMetadata.Comm))
	}
	if len(ad.Tags) > 0 {
		for _, tag := range ad.Tags {
			if !strings.HasPrefix(tag, "container_id") {
				tags = append(tags, tag)
			}
		}
	}
	if len(tags) == 0 {
		return "empty_selector"
	}
	return strings.Join(tags, ",")
}

// SendStats sends activity dump stats
func (ad *ActivityDump) SendStats() error {
	for evtType, count := range ad.processedCount {
		tags := []string{fmt.Sprintf("event_type:%s", evtType)}
		if value := count.Swap(0); value > 0 {
			if err := ad.adm.probe.statsdClient.Count(metrics.MetricActivityDumpEventProcessed, int64(value), tags, 1.0); err != nil {
				return fmt.Errorf("couldn't send %s metric: %w", metrics.MetricActivityDumpEventProcessed, err)
			}
		}
	}

	for evtType, count := range ad.addedRuntimeCount {
		tags := []string{fmt.Sprintf("event_type:%s", evtType), fmt.Sprintf("generation_type:%s", Runtime)}
		if value := count.Swap(0); value > 0 {
			if err := ad.adm.probe.statsdClient.Count(metrics.MetricActivityDumpEventAdded, int64(value), tags, 1.0); err != nil {
				return fmt.Errorf("couldn't send %s metric: %w", metrics.MetricActivityDumpEventAdded, err)
			}
		}
	}

	for evtType, count := range ad.addedSnapshotCount {
		tags := []string{fmt.Sprintf("event_type:%s", evtType), fmt.Sprintf("generation_type:%s", Snapshot)}
		if value := count.Swap(0); value > 0 {
			if err := ad.adm.probe.statsdClient.Count(metrics.MetricActivityDumpEventAdded, int64(value), tags, 1.0); err != nil {
				return fmt.Errorf("couldn't send %s metric: %w", metrics.MetricActivityDumpEventAdded, err)
			}
		}
	}

	return nil
}

// Snapshot snapshots the processes in the activity dump to capture all the
func (ad *ActivityDump) Snapshot() error {
	ad.Lock()
	defer ad.Unlock()

	for _, pan := range ad.ProcessActivityTree {
		if err := pan.snapshot(ad); err != nil {
			return err
		}
		// iterate slowly
		time.Sleep(50 * time.Millisecond)
	}

	// try to resolve the tags now
	_ = ad.resolveTags()
	return nil
}

// ResolveTags tries to resolve the activity dump tags
func (ad *ActivityDump) ResolveTags() error {
	ad.Lock()
	defer ad.Unlock()
	return ad.resolveTags()
}

// resolveTags thread unsafe version ot ResolveTags
func (ad *ActivityDump) resolveTags() error {
	if len(ad.Tags) >= 10 || len(ad.DumpMetadata.ContainerID) == 0 {
		return nil
	}

	var err error
	ad.Tags, err = ad.adm.probe.resolvers.TagsResolver.ResolveWithErr(ad.DumpMetadata.ContainerID)
	if err != nil {
		return fmt.Errorf("failed to resolve %s: %w", ad.DumpMetadata.ContainerID, err)
	}
	return nil
}

// ToSecurityActivityDumpMessage returns a pointer to a SecurityActivityDumpMessage
func (ad *ActivityDump) ToSecurityActivityDumpMessage() *api.ActivityDumpMessage {
	var storage []*api.StorageRequestMessage
	for _, requests := range ad.StorageRequests {
		for _, request := range requests {
			storage = append(storage, request.ToStorageRequestMessage(ad.DumpMetadata.Name))
		}
	}

	return &api.ActivityDumpMessage{
		Host:    ad.Host,
		Source:  ad.Source,
		Service: ad.Service,
		Tags:    ad.Tags,
		Storage: storage,
		Metadata: &api.ActivityDumpMetadataMessage{
			AgentVersion:      ad.DumpMetadata.AgentVersion,
			AgentCommit:       ad.DumpMetadata.AgentCommit,
			KernelVersion:     ad.DumpMetadata.KernelVersion,
			LinuxDistribution: ad.DumpMetadata.LinuxDistribution,
			Name:              ad.DumpMetadata.Name,
			ProtobufVersion:   ad.DumpMetadata.ProtobufVersion,
			DifferentiateArgs: ad.DumpMetadata.DifferentiateArgs,
			Comm:              ad.DumpMetadata.Comm,
			ContainerID:       ad.DumpMetadata.ContainerID,
			Start:             ad.DumpMetadata.Start.Format(time.RFC822),
			Timeout:           ad.DumpMetadata.Timeout.String(),
			Size:              ad.DumpMetadata.Size,
			Arch:              ad.DumpMetadata.Arch,
		},
	}
}

// ToTranscodingRequestMessage returns a pointer to a TranscodingRequestMessage
func (ad *ActivityDump) ToTranscodingRequestMessage() *api.TranscodingRequestMessage {
	var storage []*api.StorageRequestMessage
	for _, requests := range ad.StorageRequests {
		for _, request := range requests {
			storage = append(storage, request.ToStorageRequestMessage(ad.DumpMetadata.Name))
		}
	}

	return &api.TranscodingRequestMessage{
		Storage: storage,
	}
}

// Encode encodes an activity dump in the provided format
func (ad *ActivityDump) Encode(format dump.StorageFormat) (*bytes.Buffer, error) {
	switch format {
	case dump.JSON:
		return ad.EncodeJSON()
	case dump.MSGP:
		return ad.EncodeMSGP()
	case dump.PROTOBUF:
		return ad.EncodeProtobuf()
	case dump.PROTOJSON:
		return ad.EncodeProtoJSON()
	case dump.DOT:
		return ad.EncodeDOT()
	case dump.Profile:
		return ad.EncodeProfile()
	default:
		return nil, fmt.Errorf("couldn't encode activity dump [%s] as [%s]: unknown format", ad.GetSelectorStr(), format)
	}
}

// EncodeJSON encodes an activity dump in the JSON format
func (ad *ActivityDump) EncodeJSON() (*bytes.Buffer, error) {
	msgpRaw, err := ad.EncodeMSGP()
	if err != nil {
		return nil, err
	}

	raw := bytes.NewBuffer(nil)
	_, err = msgp.UnmarshalAsJSON(raw, msgpRaw.Bytes())
	if err != nil {
		return nil, fmt.Errorf("couldn't encode %s: %v", dump.JSON, err)
	}
	return raw, nil
}

// EncodeMSGP encodes an activity dump in the MSGP format
func (ad *ActivityDump) EncodeMSGP() (*bytes.Buffer, error) {
	ad.Lock()
	defer ad.Unlock()

	raw, err := ad.MarshalMsg(nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't encode in %s: %v", dump.MSGP, err)
	}
	return bytes.NewBuffer(raw), nil
}

// EncodeProtobuf encodes an activity dump in the Protobuf format
func (ad *ActivityDump) EncodeProtobuf() (*bytes.Buffer, error) {
	ad.Lock()
	defer ad.Unlock()

	pad := activityDumpToProto(ad)
	defer pad.ReturnToVTPool()

	raw, err := pad.MarshalVT()
	if err != nil {
		return nil, fmt.Errorf("couldn't encode in %s: %v", dump.PROTOBUF, err)
	}
	return bytes.NewBuffer(raw), nil
}

// EncodeProtoJSON encodes an activity dump in the ProtoJSON format
func (ad *ActivityDump) EncodeProtoJSON() (*bytes.Buffer, error) {
	ad.Lock()
	defer ad.Unlock()

	pad := activityDumpToProto(ad)
	defer pad.ReturnToVTPool()

	opts := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	raw, err := opts.Marshal(pad)
	if err != nil {
		return nil, fmt.Errorf("couldn't encode in %s: %v", dump.PROTOJSON, err)
	}
	return bytes.NewBuffer(raw), nil
}

// Unzip decompresses a compressed input file
func (ad *ActivityDump) Unzip(inputFile string, ext string) (string, error) {
	// uncompress the file first
	f, err := os.Open(inputFile)
	if err != nil {
		return "", fmt.Errorf("couldn't open input file: %w", err)
	}
	defer f.Close()

	seclog.Infof("unzipping %s", inputFile)
	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("couldn't create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	outputFile, err := os.Create(strings.TrimSuffix(inputFile, ext))
	if err != nil {
		return "", fmt.Errorf("couldn't create gzip output file: %w", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, gzipReader)
	if err != nil {
		return "", fmt.Errorf("couldn't unzip %s: %w", inputFile, err)
	}
	return strings.TrimSuffix(inputFile, ext), nil
}

// Decode decodes an activity dump from a file
func (ad *ActivityDump) Decode(inputFile string) error {
	var err error
	ext := filepath.Ext(inputFile)

	if ext == ".gz" {
		inputFile, err = ad.Unzip(inputFile, ext)
		if err != nil {
			return err
		}
		ext = filepath.Ext(inputFile)
	}

	format, err := dump.ParseStorageFormat(ext)
	if err != nil {
		return err
	}

	f, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("couldn't open activity dump file: %w", err)
	}
	defer f.Close()

	return ad.DecodeFromReader(f, format)
}

// DecodeFromReader decodes an activity dump from a reader with the provided format
func (ad *ActivityDump) DecodeFromReader(reader io.Reader, format dump.StorageFormat) error {
	switch format {
	case dump.MSGP:
		return ad.DecodeMSGP(reader)
	case dump.PROTOBUF:
		return ad.DecodeProtobuf(reader)
	default:
		return fmt.Errorf("unsupported input format: %s", format)
	}
}

// DecodeMSGP decodes an activity dump as MSGP
func (ad *ActivityDump) DecodeMSGP(reader io.Reader) error {
	ad.Lock()
	defer ad.Unlock()

	msgpReader := msgp.NewReader(reader)
	if err := ad.DecodeMsg(msgpReader); err != nil {
		return fmt.Errorf("couldn't parse activity dump file: %w", err)
	}
	return nil
}

// DecodeProtobuf decodes an activity dump as PROTOBUF
func (ad *ActivityDump) DecodeProtobuf(reader io.Reader) error {
	ad.Lock()
	defer ad.Unlock()

	raw, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("couldn't open activity dump file: %w", err)
	}

	inter := &adproto.ActivityDump{}
	if err := inter.UnmarshalVT(raw); err != nil {
		return fmt.Errorf("couldn't decode protobuf activity dump file: %w", err)
	}

	protoToActivityDump(ad, inter)

	return nil
}

// ProcessActivityNode holds the activity of a process
type ProcessActivityNode struct {
	id             NodeID
	Process        model.Process      `msg:"process"`
	GenerationType NodeGenerationType `msg:"generation_type"`

	Files    map[string]*FileActivityNode `msg:"files,omitempty"`
	DNSNames map[string]*DNSNode          `msg:"dns,omitempty"`
	Sockets  []*SocketNode                `msg:"sockets,omitempty"`
	Syscalls []int                        `msg:"syscalls,omitempty"`
	Children []*ProcessActivityNode       `msg:"children,omitempty"`
}

// GetID returns a unique ID to identify the current node
func (pan *ProcessActivityNode) GetID() NodeID {
	if pan.id == 0 {
		pan.id = NewNodeID()
	}
	return pan.id
}

// NewProcessActivityNode returns a new ProcessActivityNode instance
func NewProcessActivityNode(entry *model.ProcessCacheEntry, generationType NodeGenerationType) *ProcessActivityNode {
	pan := ProcessActivityNode{
		Process:        entry.Process,
		GenerationType: generationType,
		Files:          make(map[string]*FileActivityNode),
		DNSNames:       make(map[string]*DNSNode),
	}
	_ = pan.GetID()
	pan.retain()
	return &pan
}

// nolint: unused
func (pan *ProcessActivityNode) debug(prefix string) {
	fmt.Printf("%s- process: %s\n", prefix, pan.Process.FileEvent.PathnameStr)
	if len(pan.Files) > 0 {
		fmt.Printf("%s  files:\n", prefix)
		for _, f := range pan.Files {
			f.debug(fmt.Sprintf("%s\t-", prefix))
		}
	}
	if len(pan.Children) > 0 {
		fmt.Printf("%s  children:\n", prefix)
		for _, child := range pan.Children {
			child.debug(prefix + "\t")
		}
	}
}

func (pan *ProcessActivityNode) retain() {
	if pan.Process.ArgsEntry != nil && pan.Process.ArgsEntry.ArgsEnvsCacheEntry != nil {
		pan.Process.ArgsEntry.Retain()
	}
	if pan.Process.EnvsEntry != nil && pan.Process.EnvsEntry.ArgsEnvsCacheEntry != nil {
		pan.Process.EnvsEntry.Retain()
	}
}

// scrubAndReleaseArgsEnvs scrubs the process args and envs, and then releases them
func (pan *ProcessActivityNode) scrubAndReleaseArgsEnvs(resolver *ProcessResolver) {
	_, _ = resolver.GetProcessScrubbedArgv(&pan.Process)
	envs, envsTruncated := resolver.GetProcessEnvs(&pan.Process)
	pan.Process.Envs = envs
	pan.Process.EnvsTruncated = envsTruncated
	pan.Process.Argv0, _ = resolver.GetProcessArgv0(&pan.Process)

	if pan.Process.ArgsEntry != nil && pan.Process.ArgsEntry.ArgsEnvsCacheEntry != nil {
		pan.Process.ArgsEntry.Release()
	}
	if pan.Process.EnvsEntry != nil && pan.Process.EnvsEntry.ArgsEnvsCacheEntry != nil {
		pan.Process.EnvsEntry.Release()
	}
	pan.Process.ArgsEntry = nil
	pan.Process.EnvsEntry = nil
}

// Matches return true if the process fields used to generate the dump are identical with the provided ProcessCacheEntry
func (pan *ProcessActivityNode) Matches(entry *model.ProcessCacheEntry, matchArgs bool, resolvers *Resolvers) bool {

	if pan.Process.Comm == entry.Comm && pan.Process.FileEvent.PathnameStr == entry.FileEvent.PathnameStr &&
		pan.Process.Credentials == entry.Credentials {

		if matchArgs {

			panArgs, _ := resolvers.ProcessResolver.GetProcessArgv(&pan.Process)
			entryArgs, _ := resolvers.ProcessResolver.GetProcessArgv(&entry.Process)
			if len(panArgs) != len(entryArgs) {
				return false
			}

			var found bool
			for _, arg1 := range panArgs {
				found = false
				for _, arg2 := range entryArgs {
					if arg1 == arg2 {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			return true
		}

		return true
	}
	return false
}

func extractFirstParent(path string) (string, int) {
	if len(path) == 0 {
		return "", 0
	}
	if path == "/" {
		return "", 0
	}

	var add int
	if path[0] == '/' {
		path = path[1:]
		add = 1
	}

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			return path[0:i], i + add
		}
	}

	return path, len(path) + add
}

// InsertFileEvent inserts the provided file event in the current node. This function returns true if a new entry was
// added, false if the event was dropped.
func (pan *ProcessActivityNode) InsertFileEvent(fileEvent *model.FileEvent, event *Event, generationType NodeGenerationType, pathMerge bool) bool {
	parent, nextParentIndex := extractFirstParent(event.ResolveFilePath(fileEvent))
	if nextParentIndex == 0 {
		return false
	}

	// TODO: look for patterns / merge algo

	child, ok := pan.Files[parent]
	if ok {
		return child.InsertFileEvent(fileEvent, event, fileEvent.PathnameStr[nextParentIndex:], generationType, pathMerge)
	}

	// create new child
	if len(fileEvent.PathnameStr) <= nextParentIndex+1 {
		pan.Files[parent] = NewFileActivityNode(fileEvent, event, parent, generationType)
	} else {
		child := NewFileActivityNode(nil, nil, parent, generationType)
		child.InsertFileEvent(fileEvent, event, fileEvent.PathnameStr[nextParentIndex:], generationType, pathMerge)
		pan.Files[parent] = child
	}
	return true
}

// snapshot uses procfs to retrieve information about the current process
func (pan *ProcessActivityNode) snapshot(ad *ActivityDump) error {
	// call snapshot for all the children of the current node
	for _, child := range pan.Children {
		if err := child.snapshot(ad); err != nil {
			return err
		}
		// iterate slowly
		time.Sleep(50 * time.Millisecond)
	}

	// snapshot the current process
	p, err := process.NewProcess(int32(pan.Process.Pid))
	if err != nil {
		// the process doesn't exist anymore, ignore
		return nil
	}

	for _, eventType := range ad.adm.probe.config.ActivityDumpTracedEventTypes {
		switch eventType {
		case model.FileOpenEventType:
			if err = pan.snapshotFiles(p, ad); err != nil {
				return err
			}
		case model.BindEventType:
			if err = pan.snapshotBoundSockets(p, ad); err != nil {
				return err
			}
		}
	}
	return nil
}

func (pan *ProcessActivityNode) insertSnapshotedSocket(p *process.Process, ad *ActivityDump,
	family uint16, ip net.IP, port uint16) {
	evt := NewEvent(ad.adm.probe.resolvers, ad.adm.probe.scrubber, ad.adm.probe)
	evt.Event.Type = uint32(model.BindEventType)

	evt.Bind.SyscallEvent.Retval = 0
	evt.Bind.AddrFamily = family
	evt.Bind.Addr.IPNet.IP = ip
	if family == unix.AF_INET {
		evt.Bind.Addr.IPNet.Mask = net.CIDRMask(32, 32)
	} else {
		evt.Bind.Addr.IPNet.Mask = net.CIDRMask(128, 128)
	}
	evt.Bind.Addr.Port = port

	if pan.InsertBindEvent(&evt.Bind) {
		// count this new entry
		ad.addedSnapshotCount[model.BindEventType].Inc()
	}
}

func (pan *ProcessActivityNode) snapshotBoundSockets(p *process.Process, ad *ActivityDump) error {
	// list all the file descriptors opened by the process
	FDs, err := p.OpenFiles()
	if err != nil {
		return err
	}

	// sockets have the following pattern "socket:[inode]"
	var sockets []uint64
	for _, fd := range FDs {
		if strings.HasPrefix(fd.Path, "socket:[") {
			sock, err := strconv.Atoi(strings.TrimPrefix(fd.Path[:len(fd.Path)-1], "socket:["))
			if err != nil {
				return err
			}
			if sock < 0 {
				continue
			}
			sockets = append(sockets, uint64(sock))
		}
	}
	if len(sockets) <= 0 {
		return nil
	}

	// use /proc/[pid]/net/tcp,tcp6,udp,udp6 to extract the ports opened by the current process
	proc, _ := procfs.NewFS(filepath.Join(util.HostProc(fmt.Sprintf("%d", p.Pid))))
	if err != nil {
		return err
	}
	// looking for AF_INET sockets
	TCP, err := proc.NetTCP()
	if err != nil {
		seclog.Errorf("couldn't snapshot TCP sockets for [%s]: %v", ad.getSelectorStr(), err)
	}
	UDP, err := proc.NetUDP()
	if err != nil {
		seclog.Errorf("couldn't snapshot UDP sockets for [%s]: %v", ad.getSelectorStr(), err)
	}
	// looking for AF_INET6 sockets
	TCP6, err := proc.NetTCP6()
	if err != nil {
		seclog.Errorf("couldn't snapshot TCP6 sockets for [%s]: %v", ad.getSelectorStr(), err)
	}
	UDP6, err := proc.NetUDP6()
	if err != nil {
		seclog.Errorf("couldn't snapshot UDP6 sockets for [%s]: %v", ad.getSelectorStr(), err)
	}

	// searching for socket inode
	for _, s := range sockets {
		for _, sock := range TCP {
			if sock.Inode == s {
				pan.insertSnapshotedSocket(p, ad, unix.AF_INET, sock.LocalAddr,
					uint16(sock.LocalPort))
				break
			}
		}
		for _, sock := range UDP {
			if sock.Inode == s {
				pan.insertSnapshotedSocket(p, ad, unix.AF_INET, sock.LocalAddr,
					uint16(sock.LocalPort))
				break
			}
		}
		for _, sock := range TCP6 {
			if sock.Inode == s {
				pan.insertSnapshotedSocket(p, ad, unix.AF_INET6, sock.LocalAddr,
					uint16(sock.LocalPort))
				break
			}
		}
		for _, sock := range UDP6 {
			if sock.Inode == s {
				pan.insertSnapshotedSocket(p, ad, unix.AF_INET6, sock.LocalAddr,
					uint16(sock.LocalPort))
				break
			}
		}
		// not necessary found here, can be also another kind of socket (AF_UNIX, AF_NETLINK, etc)
	}
	return nil
}

func (pan *ProcessActivityNode) snapshotFiles(p *process.Process, ad *ActivityDump) error {
	// list the files opened by the process
	fileFDs, err := p.OpenFiles()
	if err != nil {
		return err
	}

	var files []string
	for _, fd := range fileFDs {
		files = append(files, fd.Path)
	}

	// list the mmaped files of the process
	memoryMaps, err := p.MemoryMaps(false)
	if err != nil {
		return err
	}

	for _, mm := range *memoryMaps {
		if mm.Path != pan.Process.FileEvent.PathnameStr {
			files = append(files, mm.Path)
		}
	}

	// insert files
	var fileinfo os.FileInfo
	var resolvedPath string
	for _, f := range files {
		if len(f) == 0 {
			continue
		}

		// fetch the file user, group and mode
		fullPath := filepath.Join(utils.RootPath(int32(pan.Process.Pid)), f)
		fileinfo, err = os.Stat(fullPath)
		if err != nil {
			continue
		}
		stat, ok := fileinfo.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}

		evt := NewEvent(ad.adm.probe.resolvers, ad.adm.probe.scrubber, ad.adm.probe)
		evt.Event.Type = uint32(model.FileOpenEventType)

		resolvedPath, err = filepath.EvalSymlinks(f)
		if err != nil {
			evt.Open.File.PathnameStr = resolvedPath
		} else {
			evt.Open.File.PathnameStr = f
		}
		evt.Open.File.BasenameStr = path.Base(evt.Open.File.PathnameStr)
		evt.Open.File.FileFields.Mode = uint16(stat.Mode)
		evt.Open.File.FileFields.Inode = stat.Ino
		evt.Open.File.FileFields.UID = stat.Uid
		evt.Open.File.FileFields.GID = stat.Gid
		evt.Open.File.FileFields.MTime = uint64(ad.adm.probe.resolvers.TimeResolver.ComputeMonotonicTimestamp(time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)))
		evt.Open.File.FileFields.CTime = uint64(ad.adm.probe.resolvers.TimeResolver.ComputeMonotonicTimestamp(time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec)))

		evt.Open.File.Mode = evt.Open.File.FileFields.Mode
		// TODO: add open flags by parsing `/proc/[pid]/fdinfo/fd` + O_RDONLY|O_CLOEXEC for the shared libs

		if pan.InsertFileEvent(&evt.Open.File, evt, Snapshot, ad.adm.probe.config.ActivityDumpPathMergeEnabled) {
			// count this new entry
			ad.addedSnapshotCount[model.FileOpenEventType].Inc()
		}
	}
	return nil
}

// InsertDNSEvent inserts
func (pan *ProcessActivityNode) InsertDNSEvent(evt *model.DNSEvent) bool {
	if dnsNode, ok := pan.DNSNames[evt.Name]; ok {
		// look for the DNS request type
		for _, req := range dnsNode.Requests {
			if req.Type == evt.Type {
				return false
			}
		}

		// insert the new request
		dnsNode.Requests = append(dnsNode.Requests, *evt)
		return true
	}
	pan.DNSNames[evt.Name] = NewDNSNode(evt)
	return true
}

// InsertBindEvent inserts a bind event to the activity dump
func (pan *ProcessActivityNode) InsertBindEvent(evt *model.BindEvent) bool {
	if evt.SyscallEvent.Retval != 0 {
		return false
	}
	var newNode bool
	evtFamily := model.AddressFamily(evt.AddrFamily).String()

	// check if a socket of this type already exists
	var sock *SocketNode
	for _, s := range pan.Sockets {
		if s.Family == evtFamily {
			sock = s
		}
	}
	if sock == nil {
		sock = NewSocketNode(evtFamily)
		pan.Sockets = append(pan.Sockets, sock)
		newNode = true
	}

	// Insert bind event
	if sock.InsertBindEvent(evt) {
		newNode = true
	}

	return newNode
}

// InsertSyscalls inserts the syscall of the process in the dump
func (pan *ProcessActivityNode) InsertSyscalls(e *model.SyscallsEvent) bool {
	var hasNewSyscalls bool
newSyscallLoop:
	for _, newSyscall := range e.Syscalls {
		for _, existingSyscall := range pan.Syscalls {
			if existingSyscall == int(newSyscall) {
				continue newSyscallLoop
			}
		}

		pan.Syscalls = append(pan.Syscalls, int(newSyscall))
		hasNewSyscalls = true
	}
	return hasNewSyscalls
}

// FileActivityNode holds a tree representation of a list of files
type FileActivityNode struct {
	id             NodeID
	Name           string             `msg:"name"`
	File           *model.FileEvent   `msg:"file,omitempty"`
	GenerationType NodeGenerationType `msg:"generation_type"`
	FirstSeen      time.Time          `msg:"first_seen,omitempty"`

	Open *OpenNode `msg:"open,omitempty"`

	Children map[string]*FileActivityNode `msg:"children,omitempty"`
}

// GetID returns a unique ID to identify the current node
func (fan *FileActivityNode) GetID() NodeID {
	if fan.id == 0 {
		fan.id = NewNodeID()
	}
	return fan.id
}

// OpenNode contains the relevant fields of an Open event on which we might want to write a profiling rule
type OpenNode struct {
	model.SyscallEvent
	Flags uint32 `msg:"flags"`
	Mode  uint32 `msg:"mode"`
}

// NewFileActivityNode returns a new FileActivityNode instance
func NewFileActivityNode(fileEvent *model.FileEvent, event *Event, name string, generationType NodeGenerationType) *FileActivityNode {
	fan := &FileActivityNode{
		Name:           name,
		GenerationType: generationType,
		Children:       make(map[string]*FileActivityNode),
	}
	_ = fan.GetID()
	if fileEvent != nil {
		fileEventTmp := *fileEvent
		fan.File = &fileEventTmp
	}
	fan.enrichFromEvent(event)
	return fan
}

func (fan *FileActivityNode) getNodeLabel() string {
	label := fan.Name
	if fan.Open != nil {
		label += " [open]"
	}
	return label
}

func (fan *FileActivityNode) enrichFromEvent(event *Event) {
	if event == nil {
		return
	}
	if fan.FirstSeen.IsZero() {
		fan.FirstSeen = event.ResolveEventTimestamp()
	}

	switch event.GetEventType() {
	case model.FileOpenEventType:
		fan.Open = &OpenNode{
			SyscallEvent: event.Open.SyscallEvent,
			Flags:        event.Open.Flags,
			Mode:         event.Open.Mode,
		}
	}
}

// InsertFileEvent inserts an event in a FileActivityNode. This function returns true if a new entry was added, false if
// the event was dropped.
func (fan *FileActivityNode) InsertFileEvent(fileEvent *model.FileEvent, event *Event, remainingPath string, generationType NodeGenerationType, pathMerge bool) bool {
	parent, nextParentIndex := extractFirstParent(remainingPath)
	if nextParentIndex == 0 {
		fan.enrichFromEvent(event)
		return false
	}

	if pathMerge && len(fan.Children) >= 10 {
		fan.mergeCommonPaths()
	}

	child, ok := fan.Children[parent]
	if ok {
		return child.InsertFileEvent(fileEvent, event, remainingPath[nextParentIndex:], generationType, pathMerge)
	}

	// create new child
	if len(remainingPath) <= nextParentIndex+1 {
		fan.Children[parent] = NewFileActivityNode(fileEvent, event, parent, generationType)
	} else {
		child := NewFileActivityNode(nil, nil, parent, generationType)
		child.InsertFileEvent(fileEvent, event, remainingPath[nextParentIndex:], generationType, pathMerge)
		fan.Children[parent] = child
	}
	return true
}

func (fan *FileActivityNode) mergeCommonPaths() {
	fan.Children = combineChildren(fan.Children)
}

func combineChildren(children map[string]*FileActivityNode) map[string]*FileActivityNode {
	if len(children) == 0 {
		return children
	}

	type inner struct {
		pair utils.StringPair
		fan  *FileActivityNode
	}

	inputs := make([]inner, 0, len(children))
	for k, v := range children {
		inputs = append(inputs, inner{
			pair: utils.NewStringPair(k),
			fan:  v,
		})
	}

	current := []inner{inputs[0]}

	for _, a := range inputs[1:] {
		next := make([]inner, 0)
		shouldAppend := true
		for _, b := range current {
			if !areCompatibleFans(a.fan, b.fan) {
				next = append(next, b)
				continue
			}

			sp, similar := utils.BuildGlob(a.pair, b.pair, 4)
			if similar {
				merged, ok := mergeFans(sp.ToGlob(), a.fan, b.fan)
				if !ok {
					next = append(next, b)
					continue
				}

				next = append(next, inner{
					pair: sp,
					fan:  merged,
				})
				shouldAppend = false
			}
		}

		if shouldAppend {
			next = append(next, a)
		}
		current = next
	}

	res := make(map[string]*FileActivityNode)
	for _, n := range current {
		glob := n.pair.ToGlob()
		n.fan.Name = glob
		res[glob] = n.fan
	}

	return res
}

func areCompatibleFans(a *FileActivityNode, b *FileActivityNode) bool {
	return reflect.DeepEqual(a.Open, b.Open)
}

func mergeFans(name string, a *FileActivityNode, b *FileActivityNode) (*FileActivityNode, bool) {
	newChildren := make(map[string]*FileActivityNode)
	for k, v := range a.Children {
		newChildren[k] = v
	}
	for k, v := range b.Children {
		if _, present := newChildren[k]; present {
			return nil, false
		}
		newChildren[k] = v
	}

	return &FileActivityNode{
		id:             a.id,
		Name:           name,
		File:           a.File,
		GenerationType: a.GenerationType,
		FirstSeen:      a.FirstSeen,
		Open:           a.Open, // if the 2 fans are compatible, a.Open should be equal to b.Open
		Children:       newChildren,
	}, true
}

// nolint: unused
func (fan *FileActivityNode) debug(prefix string) {
	fmt.Printf("%s %s\n", prefix, fan.Name)
	for _, child := range fan.Children {
		child.debug("\t" + prefix)
	}
}

// DNSNode is used to store a DNS node
type DNSNode struct {
	id       NodeID
	Requests []model.DNSEvent `msg:"requests"`
}

// NewDNSNode returns a new DNSNode instance
func NewDNSNode(event *model.DNSEvent) *DNSNode {
	return &DNSNode{
		Requests: []model.DNSEvent{*event},
	}
}

// GetID returns the ID of the current DNS node
func (n *DNSNode) GetID() NodeID {
	if n.id == 0 {
		n.id = NewNodeID()
	}
	return n.id
}

// BindNode is used to store a bind node
type BindNode struct {
	Port uint16 `msg:"port"`
	IP   string `msg:"ip"`
}

// SocketNode is used to store a Socket node and associated events
type SocketNode struct {
	id     NodeID
	Family string      `msg:"family"`
	Bind   []*BindNode `msg:"bind,omitempty"`
}

// InsertBindEvent inserts a bind even inside a socket node
func (n *SocketNode) InsertBindEvent(evt *model.BindEvent) bool {
	// ignore non IPv4 / IPv6 bind events for now
	if evt.AddrFamily != unix.AF_INET && evt.AddrFamily != unix.AF_INET6 {
		return false
	}
	evtIP := evt.Addr.IPNet.IP.String()

	for _, n := range n.Bind {
		if evt.Addr.Port == n.Port && evtIP == n.IP {
			return false
		}
	}

	// insert bind event now
	n.Bind = append(n.Bind, &BindNode{
		Port: evt.Addr.Port,
		IP:   evtIP,
	})
	return true
}

// NewSocketNode returns a new SocketNode instance
func NewSocketNode(family string) *SocketNode {
	return &SocketNode{
		Family: family,
	}
}

// GetID returns the ID of the current Socket node
func (n *SocketNode) GetID() NodeID {
	if n.id == 0 {
		n.id = NewNodeID()
	}
	return n.id
}

// NodeID represents the ID of a Node
//msgp:ignore NodeID
type NodeID uint64

// NewNodeID returns a new random NodeID
func NewNodeID() NodeID {
	return NodeID(eval.RandNonZeroUint64())
}

// IsUnset checks if the NodeID is unset
func (id NodeID) IsUnset() bool {
	return id == 0
}

func (id NodeID) String() string {
	return fmt.Sprintf("node%d", uint64(id))
}
