package output

// Fork of github.com/moriyoshi/pulsego with a bunch of modifications.
// TODO(padster): Move these upstream and use the moriyoshi version instead.

/*
#cgo LDFLAGS: -lpulse
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <pulse/error.h>
#include <pulse/context.h>
#include <pulse/stream.h>
#include <pulse/thread-mainloop.h>

typedef struct {
    pa_threaded_mainloop *pa;
    int done;
    int status;
} state_cb_param_t;

static void context_on_completion(pa_context *ctx, int status, state_cb_param_t *param)
{
    param->status = status;
    param->done = 1;
    pa_threaded_mainloop_signal(param->pa, 0);
}

static void context_on_state_change(pa_context *ctx, pa_threaded_mainloop *pa)
{
    pa_threaded_mainloop_signal(pa, 0);
}

static void stream_on_state_change(pa_stream *st, pa_threaded_mainloop *pa)
{
    pa_threaded_mainloop_signal(pa, 0);
}

static void stream_on_success(pa_stream *st, int success, pa_threaded_mainloop *pa)
{
    pa_threaded_mainloop_signal(pa, 0);
}

int context_poll_unless(pa_threaded_mainloop *pa, pa_context *ctx, pa_context_state_t state)
{
    pa_context_state_t s;
    pa_threaded_mainloop_lock(pa);
    for (;;) {
        s = pa_context_get_state(ctx);
        if (s == state || s == PA_CONTEXT_FAILED || s == PA_CONTEXT_TERMINATED)
            break;
        pa_threaded_mainloop_wait(pa);
    }
    pa_threaded_mainloop_unlock(pa);
    return s;
}

int context_poll_until_done(state_cb_param_t *param, pa_context *ctx)
{
    for (;;) {
        if (param->done)
            break;
        pa_threaded_mainloop_wait(param->pa);
    }
    return param->status;
}

int stream_poll_unless(pa_threaded_mainloop *pa, pa_stream *st, pa_stream_state_t state)
{
    pa_stream_state_t s;
    pa_threaded_mainloop_lock(pa);
    for (;;) {
        s = pa_stream_get_state(st);
        if (s == state || s == PA_STREAM_FAILED || s == PA_STREAM_TERMINATED)
            break;
        pa_threaded_mainloop_wait(pa);
    }
    pa_threaded_mainloop_unlock(pa);
    return s;
}

int context_set_default_sink(pa_threaded_mainloop *pa, pa_context *ctx, const char *name)
{
    pa_operation *op;
    state_cb_param_t param = { pa, 0, 0 };

    pa_threaded_mainloop_lock(pa);
    op = pa_context_set_default_sink(ctx, name, (pa_context_success_cb_t)context_on_completion, pa);
    pa_threaded_mainloop_unlock(pa);
    context_poll_until_done(&param, ctx);
    pa_operation_unref(op);
    return param.status;
}

int context_set_default_source(pa_threaded_mainloop *pa, pa_context *ctx, const char *name)
{
    pa_operation *op;
    state_cb_param_t param = { pa, 0, 0 };

    pa_threaded_mainloop_lock(pa);
    op = pa_context_set_default_source(ctx, name, (pa_context_success_cb_t)context_on_completion, pa);
    pa_threaded_mainloop_unlock(pa);
    context_poll_until_done(&param, ctx);
    pa_operation_unref(op);
    return param.status;
}

int context_exit_daemon(pa_threaded_mainloop *pa, pa_context *ctx)
{
    pa_operation *op;
    state_cb_param_t param = { pa, 0, 0 };

    pa_threaded_mainloop_lock(pa);
    op = pa_context_exit_daemon(ctx, (pa_context_success_cb_t)context_on_completion, pa);
    pa_threaded_mainloop_unlock(pa);
    context_poll_until_done(&param, ctx);
    if (op)
        pa_operation_unref(op);
    return param.status;
}

int context_drain(pa_threaded_mainloop *pa, pa_context *ctx)
{
    pa_operation *op;
    state_cb_param_t param = { pa, 0, 0 };

    pa_threaded_mainloop_lock(pa);
    op = pa_context_drain(ctx, (pa_context_notify_cb_t)context_on_state_change, pa);
    pa_threaded_mainloop_unlock(pa);
    pa_threaded_mainloop_wait(pa);
    if (op)
        pa_operation_unref(op);
    return param.status;
}

pa_context *context_new(pa_threaded_mainloop *pa, const char *name, pa_context_flags_t flags)
{
    pa_mainloop_api *api = pa_threaded_mainloop_get_api(pa);
    pa_context *ctx = pa_context_new(api, name);
    int err = PA_OK;

    pa_context_set_state_callback(ctx,
            (pa_context_notify_cb_t)context_on_state_change, pa);

    {
        pa_threaded_mainloop_lock(pa);
        err = pa_context_connect(ctx, NULL, flags, NULL);
        pa_threaded_mainloop_unlock(pa);
        if (err < 0)
            return NULL;
    }

    if (context_poll_unless(pa, ctx, PA_CONTEXT_READY) != PA_CONTEXT_READY) {
        pa_context_unref(ctx);
        return NULL;
    }
    return ctx;
}

pa_stream *stream_new(pa_threaded_mainloop *pa, pa_context *ctx, const char *name, pa_sample_format_t format, int rate, int channels)
{
    pa_stream *st = NULL;
    {
        pa_threaded_mainloop_lock(pa);
        {
            pa_sample_spec ss;
            ss.format = format;
            ss.rate = rate;
            ss.channels = channels;
            st = pa_stream_new(ctx, name, &ss, NULL);
            if (!st) {
                pa_threaded_mainloop_unlock(pa);
                return NULL;
            }
        }

        pa_stream_set_state_callback(st,
                (pa_stream_notify_cb_t)stream_on_state_change, pa);
        pa_threaded_mainloop_unlock(pa);
    }

{
    int s = context_poll_unless(pa, ctx, PA_CONTEXT_READY);
    if (s != PA_CONTEXT_READY) {
        pa_stream_unref(st);
        return NULL;
    }
}
    return st;
}

int stream_write(pa_threaded_mainloop *pa, pa_stream *p, const void *data, size_t nbytes, pa_seek_mode_t seek)
{
    int retval;
    pa_threaded_mainloop_lock(pa);
    retval = pa_stream_write(p, data, nbytes, NULL, 0, seek);
    pa_threaded_mainloop_unlock(pa);
    return retval;
}

// Added: Read read the available size to be written
size_t stream_writable_size(pa_stream *p)
{
    return pa_stream_writable_size(p);
}

// Added: Drain a stream, and wait for it to finish before returning.
void stream_drain(pa_threaded_mainloop *pa, pa_stream *s)
{
    pa_operation *op;
    pa_threaded_mainloop_lock(pa);
    op = pa_stream_drain(s, (pa_stream_success_cb_t)stream_on_success, pa);
    while (pa_operation_get_state(op) == PA_OPERATION_RUNNING)
        pa_threaded_mainloop_wait(pa);

    pa_threaded_mainloop_unlock(pa);
    pa_threaded_mainloop_wait(pa);
    pa_operation_unref(op);
}
*/
import "C"
import (
	"reflect"
	"unsafe"
)

const (
	CONTEXT_UNCONNECTED  = C.PA_CONTEXT_UNCONNECTED
	CONTEXT_CONNECTING   = C.PA_CONTEXT_CONNECTING
	CONTEXT_AUTHORIZING  = C.PA_CONTEXT_AUTHORIZING
	CONTEXT_SETTING_NAME = C.PA_CONTEXT_SETTING_NAME
	CONTEXT_READY        = C.PA_CONTEXT_READY
	CONTEXT_FAILED       = C.PA_CONTEXT_FAILED
	CONTEXT_TERMINATED   = C.PA_CONTEXT_TERMINATED
)

const (
	STREAM_UNCONNECTED = C.PA_STREAM_UNCONNECTED
	STREAM_CREATING    = C.PA_STREAM_CREATING
	STREAM_READY       = C.PA_STREAM_READY
	STREAM_FAILED      = C.PA_STREAM_FAILED
	STREAM_TERMINATED  = C.PA_STREAM_TERMINATED
)

/*
const (
    OPERATION_RUNNING = C.PA_OPERATION_RUNNING;
    OPERATION_DONE = C.PA_OPERATION_DONE;
    OPERATION_CANCELLED = C.PA_OPERATION_CANCELLED
)
*/

const (
	CONTEXT_NOFLAGS     = C.PA_CONTEXT_NOFLAGS
	CONTEXT_NOAUTOSPAWN = C.PA_CONTEXT_NOAUTOSPAWN
	CONTEXT_NOFAIL      = C.PA_CONTEXT_NOFAIL
)

/*
const (
    STREAM_NODIRECTION = C.PA_STREAM_NODIRECTION;
    STREAM_PLAYBACK = C.PA_STREAM_PLAYBACK;
    STREAM_RECORD = C.PA_STREAM_RECORD;
    STREAM_UPLOAD = C.PA_STREAM_UPLOAD
)
*/

const (
	STREAM_NOFLAGS                   = C.PA_STREAM_NOFLAGS
	STREAM_START_CORKED              = C.PA_STREAM_START_CORKED
	STREAM_INTERPOLATE_TIMING        = C.PA_STREAM_INTERPOLATE_TIMING
	STREAM_NOT_MONOTONIC             = C.PA_STREAM_NOT_MONOTONIC
	STREAM_AUTO_TIMING_UPDATE        = C.PA_STREAM_AUTO_TIMING_UPDATE
	STREAM_NO_REMAP_CHANNELS         = C.PA_STREAM_NO_REMAP_CHANNELS
	STREAM_NO_REMIX_CHANNELS         = C.PA_STREAM_NO_REMIX_CHANNELS
	STREAM_FIX_FORMAT                = C.PA_STREAM_FIX_FORMAT
	STREAM_FIX_RATE                  = C.PA_STREAM_FIX_RATE
	STREAM_FIX_CHANNELS              = C.PA_STREAM_FIX_CHANNELS
	STREAM_DONT_MOVE                 = C.PA_STREAM_DONT_MOVE
	STREAM_VARIABLE_RATE             = C.PA_STREAM_VARIABLE_RATE
	STREAM_PEAK_DETECT               = C.PA_STREAM_PEAK_DETECT
	STREAM_START_MUTED               = C.PA_STREAM_START_MUTED
	STREAM_ADJUST_LATENCY            = C.PA_STREAM_ADJUST_LATENCY
	STREAM_EARLY_REQUESTS            = C.PA_STREAM_EARLY_REQUESTS
	STREAM_DONT_INHIBIT_AUTO_SUSPEND = C.PA_STREAM_DONT_INHIBIT_AUTO_SUSPEND
	STREAM_START_UNMUTED             = C.PA_STREAM_START_UNMUTED
	STREAM_FAIL_ON_SUSPEND           = C.PA_STREAM_FAIL_ON_SUSPEND
)

const (
	OK                       = C.PA_OK
	ERR_ACCESS               = C.PA_ERR_ACCESS
	ERR_COMMAND              = C.PA_ERR_COMMAND
	ERR_INVALID              = C.PA_ERR_INVALID
	ERR_EXIST                = C.PA_ERR_EXIST
	ERR_NOENTITY             = C.PA_ERR_NOENTITY
	ERR_CONNECTIONREFUSED    = C.PA_ERR_CONNECTIONREFUSED
	ERR_PROTOCOL             = C.PA_ERR_PROTOCOL
	ERR_TIMEOUT              = C.PA_ERR_TIMEOUT
	ERR_AUTHKEY              = C.PA_ERR_AUTHKEY
	ERR_INTERNAL             = C.PA_ERR_INTERNAL
	ERR_CONNECTIONTERMINATED = C.PA_ERR_CONNECTIONTERMINATED
	ERR_KILLED               = C.PA_ERR_KILLED
	ERR_INVALIDSERVER        = C.PA_ERR_INVALIDSERVER
	ERR_MODINITFAILED        = C.PA_ERR_MODINITFAILED
	ERR_BADSTATE             = C.PA_ERR_BADSTATE
	ERR_NODATA               = C.PA_ERR_NODATA
	ERR_VERSION              = C.PA_ERR_VERSION
	ERR_TOOLARGE             = C.PA_ERR_TOOLARGE
	ERR_NOTSUPPORTED         = C.PA_ERR_NOTSUPPORTED
	ERR_UNKNOWN              = C.PA_ERR_UNKNOWN
	ERR_NOEXTENSION          = C.PA_ERR_NOEXTENSION
	ERR_OBSOLETE             = C.PA_ERR_OBSOLETE
	ERR_NOTIMPLEMENTED       = C.PA_ERR_NOTIMPLEMENTED
	ERR_FORKED               = C.PA_ERR_FORKED
	ERR_IO                   = C.PA_ERR_IO
	ERR_BUSY                 = C.PA_ERR_BUSY
	ERR_MAX                  = C.PA_ERR_MAX
)

/*
const (
    SUBSCRIPTION_MASK_NULL = C.PA_SUBSCRIPTION_MASK_NULL;
    SUBSCRIPTION_MASK_SINK = C.PA_SUBSCRIPTION_MASK_SINK;
    SUBSCRIPTION_MASK_SOURCE = C.PA_SUBSCRIPTION_MASK_SOURCE;
    SUBSCRIPTION_MASK_SINK_INPUT = C.PA_SUBSCRIPTION_MASK_SINK_INPUT;
    SUBSCRIPTION_MASK_SOURCE_OUTPUT = C.PA_SUBSCRIPTION_MASK_SOURCE_OUTPUT;
    SUBSCRIPTION_MASK_MODULE = C.PA_SUBSCRIPTION_MASK_MODULE;
    SUBSCRIPTION_MASK_CLIENT = C.PA_SUBSCRIPTION_MASK_CLIENT;
    SUBSCRIPTION_MASK_SAMPLE_CACHE = C.PA_SUBSCRIPTION_MASK_SAMPLE_CACHE;
    SUBSCRIPTION_MASK_SERVER = C.PA_SUBSCRIPTION_MASK_SERVER;
    SUBSCRIPTION_MASK_AUTOLOAD = C.PA_SUBSCRIPTION_MASK_AUTOLOAD;
    SUBSCRIPTION_MASK_CARD = C.PA_SUBSCRIPTION_MASK_CARD;
    SUBSCRIPTION_MASK_ALL = C.PA_SUBSCRIPTION_MASK_ALL
)

const (
    SUBSCRIPTION_EVENT_SINK = C.PA_SUBSCRIPTION_EVENT_SINK;
    SUBSCRIPTION_EVENT_SOURCE = C.PA_SUBSCRIPTION_EVENT_SOURCE;
    SUBSCRIPTION_EVENT_SINK_INPUT = C.PA_SUBSCRIPTION_EVENT_SINK_INPUT;
    SUBSCRIPTION_EVENT_SOURCE_OUTPUT = C.PA_SUBSCRIPTION_EVENT_SOURCE_OUTPUT;
    SUBSCRIPTION_EVENT_MODULE = C.PA_SUBSCRIPTION_EVENT_MODULE;
    SUBSCRIPTION_EVENT_CLIENT = C.PA_SUBSCRIPTION_EVENT_CLIENT;
    SUBSCRIPTION_EVENT_SAMPLE_CACHE = C.PA_SUBSCRIPTION_EVENT_SAMPLE_CACHE;
    SUBSCRIPTION_EVENT_SERVER = C.PA_SUBSCRIPTION_EVENT_SERVER;
    SUBSCRIPTION_EVENT_AUTOLOAD = C.PA_SUBSCRIPTION_EVENT_AUTOLOAD;
    SUBSCRIPTION_EVENT_CARD = C.PA_SUBSCRIPTION_EVENT_CARD;
    SUBSCRIPTION_EVENT_FACILITY_MASK = C.PA_SUBSCRIPTION_EVENT_FACILITY_MASK;
    SUBSCRIPTION_EVENT_NEW = C.PA_SUBSCRIPTION_EVENT_NEW;
    SUBSCRIPTION_EVENT_CHANGE = C.PA_SUBSCRIPTION_EVENT_CHANGE;
    SUBSCRIPTION_EVENT_REMOVE = C.PA_SUBSCRIPTION_EVENT_REMOVE;
    SUBSCRIPTION_EVENT_TYPE_MASK = C.PA_SUBSCRIPTION_EVENT_TYPE_MASK
)
*/

const (
	SEEK_RELATIVE         = C.PA_SEEK_RELATIVE
	SEEK_ABSOLUTE         = C.PA_SEEK_ABSOLUTE
	SEEK_RELATIVE_ON_READ = C.PA_SEEK_RELATIVE_ON_READ
	SEEK_RELATIVE_END     = C.PA_SEEK_RELATIVE_END
)

/*
const (
    SINK_NOFLAGS = C.PA_SINK_NOFLAGS;
    SINK_HW_VOLUME_CTRL = C.PA_SINK_HW_VOLUME_CTRL;
    SINK_LATENCY = C.PA_SINK_LATENCY;
    SINK_HARDWARE = C.PA_SINK_HARDWARE;
    SINK_NETWORK = C.PA_SINK_NETWORK;
    SINK_HW_MUTE_CTRL = C.PA_SINK_HW_MUTE_CTRL;
    SINK_DECIBEL_VOLUME = C.PA_SINK_DECIBEL_VOLUME;
    SINK_FLAT_VOLUME = C.PA_SINK_FLAT_VOLUME;
    SINK_DYNAMIC_LATENCY = C.PA_SINK_DYNAMIC_LATENCY
)
const (
    SINK_INVALID_STATE = C.PA_SINK_INVALID_STATE;
    SINK_RUNNING = C.PA_SINK_RUNNING;
    SINK_IDLE = C.PA_SINK_IDLE;
    SINK_SUSPENDED = C.PA_SINK_SUSPENDED;
    SINK_INIT = C.PA_SINK_INIT;
    SINK_UNLINKED = C.PA_SINK_UNLINKED
)

const (
    SOURCE_NOFLAGS = C.PA_SOURCE_NOFLAGS;
    SOURCE_HW_VOLUME_CTRL = C.PA_SOURCE_HW_VOLUME_CTRL;
    SOURCE_LATENCY = C.PA_SOURCE_LATENCY;
    SOURCE_HARDWARE = C.PA_SOURCE_HARDWARE;
    SOURCE_NETWORK = C.PA_SOURCE_NETWORK;
    SOURCE_HW_MUTE_CTRL = C.PA_SOURCE_HW_MUTE_CTRL;
    SOURCE_DECIBEL_VOLUME = C.PA_SOURCE_DECIBEL_VOLUME;
    SOURCE_DYNAMIC_LATENCY = C.PA_SOURCE_DYNAMIC_LATENCY
)

const (
    SOURCE_INVALID_STATE = C.PA_SOURCE_INVALID_STATE;
    SOURCE_RUNNING = C.PA_SOURCE_RUNNING;
    SOURCE_IDLE = C.PA_SOURCE_IDLE;
    SOURCE_SUSPENDED = C.PA_SOURCE_SUSPENDED;
    SOURCE_INIT = C.PA_SOURCE_INIT;
    SOURCE_UNLINKED = C.PA_SOURCE_UNLINKED
)
*/

const (
	SAMPLE_U8        = C.PA_SAMPLE_U8
	SAMPLE_ALAW      = C.PA_SAMPLE_ALAW
	SAMPLE_ULAW      = C.PA_SAMPLE_ULAW
	SAMPLE_S16LE     = C.PA_SAMPLE_S16LE
	SAMPLE_S16BE     = C.PA_SAMPLE_S16BE
	SAMPLE_FLOAT32LE = C.PA_SAMPLE_FLOAT32LE
	SAMPLE_FLOAT32BE = C.PA_SAMPLE_FLOAT32BE
	SAMPLE_S32LE     = C.PA_SAMPLE_S32LE
	SAMPLE_S32BE     = C.PA_SAMPLE_S32BE
	SAMPLE_S24LE     = C.PA_SAMPLE_S24LE
	SAMPLE_S24BE     = C.PA_SAMPLE_S24BE
	SAMPLE_S24_32LE  = C.PA_SAMPLE_S24_32LE
	SAMPLE_S24_32BE  = C.PA_SAMPLE_S24_32BE
	SAMPLE_MAX       = C.PA_SAMPLE_MAX
	SAMPLE_INVALID   = C.PA_SAMPLE_INVALID
)

type PulseMainLoop struct {
	pa *C.pa_threaded_mainloop
}

type PulseContext struct {
	MainLoop *PulseMainLoop
	ctx      *C.pa_context
}

type PulseStream struct {
	Context *PulseContext
	st      *C.pa_stream
}

type PulseSampleSpec struct {
	Format   int
	Rate     int
	Channels int
}

func (self *PulseStream) Disconnect() {
	C.pa_threaded_mainloop_lock(self.Context.MainLoop.pa)
	C.pa_stream_disconnect(self.st)
	C.pa_threaded_mainloop_unlock(self.Context.MainLoop.pa)
}

func (self *PulseStream) Dispose() {
	self.Disconnect()
	C.pa_stream_unref(self.st)
}

func (self *PulseStream) ConnectToSink() int {
	C.pa_threaded_mainloop_lock(self.Context.MainLoop.pa)
	err := C.pa_stream_connect_playback(self.st, nil, nil, 0, nil, nil)
	C.pa_threaded_mainloop_unlock(self.Context.MainLoop.pa)
	if err == OK {
		err = C.stream_poll_unless(self.Context.MainLoop.pa, self.st, STREAM_READY)
	}
	return int(err)
}

func (self *PulseStream) GetSampleSpec() PulseSampleSpec {
	spec := C.pa_stream_get_sample_spec(self.st)
	return PulseSampleSpec{
		Format:   int(spec.format),
		Rate:     int(spec.rate),
		Channels: int(spec.channels),
	}
}

func (self *PulseStream) Write(data interface{}, flags int) int {
	typ := reflect.TypeOf(data)
	format := self.GetSampleSpec().Format
	samples := reflect.ValueOf(data)
	nsamples := samples.Len()
	ptr := unsafe.Pointer(samples.Index(0).UnsafeAddr())
	switch typ.Elem().Kind() {
	case reflect.Int32:
		if format != SAMPLE_S32LE && format != SAMPLE_S24_32LE {
			return ERR_INVALID
		}
	case reflect.Float32:
		if format != SAMPLE_FLOAT32LE {
			return ERR_INVALID
		}
	case reflect.Int16:
		if format != SAMPLE_S16LE {
			return ERR_INVALID
		}
	case reflect.Uint8:
		if format != SAMPLE_U8 {
			return ERR_INVALID
		}
	}
	retval := int(C.stream_write(self.Context.MainLoop.pa, self.st, ptr, C.size_t(typ.Elem().Size()*uintptr(nsamples)), C.pa_seek_mode_t(flags)))
	return retval
}

func (self *PulseStream) WritableSize() int {
	return int(C.stream_writable_size(self.st))
}

func (self *PulseStream) Drain() {
	C.stream_drain(self.Context.MainLoop.pa, self.st)
}

func (self *PulseContext) Disconnect() {
	C.pa_threaded_mainloop_lock(self.MainLoop.pa)
	C.pa_context_disconnect(self.ctx)
	C.pa_threaded_mainloop_unlock(self.MainLoop.pa)
}

func (self *PulseContext) Dispose() {
	self.Disconnect()
	C.pa_context_unref(self.ctx)
}

func (self *PulseContext) Drain() int {
	return int(C.context_drain(self.MainLoop.pa, self.ctx))
}

func (self *PulseContext) ExitDaemon() int {
	return int(C.context_exit_daemon(self.MainLoop.pa, self.ctx))
}

func (self *PulseContext) SetDefaultSource(name string) int {
	name_ := C.CString(name)
	retval := int(C.context_set_default_source(self.MainLoop.pa, self.ctx, name_))
	C.free(unsafe.Pointer(name_))
	return retval
}

func (self *PulseContext) SetDefaultSink(name string) int {
	name_ := C.CString(name)
	retval := int(C.context_set_default_sink(self.MainLoop.pa, self.ctx, name_))
	C.free(unsafe.Pointer(name_))
	return retval
}

func (self *PulseContext) NewStream(name string, spec *PulseSampleSpec) *PulseStream {
	name_ := C.CString(name)
	st := C.stream_new(self.MainLoop.pa, self.ctx, name_, C.pa_sample_format_t(spec.Format), C.int(spec.Rate), C.int(spec.Channels))
	var retval *PulseStream = nil
	if st != nil {
		retval = &PulseStream{Context: self, st: st}
	}
	C.free(unsafe.Pointer(name_))
	return retval
}

func (self *PulseMainLoop) NewContext(name string, flags int) *PulseContext {
	name_ := C.CString(name)
	ctx := C.context_new(self.pa, name_, C.pa_context_flags_t(flags))
	var retval *PulseContext = nil
	if ctx != nil {
		retval = &PulseContext{MainLoop: self, ctx: ctx}
	}
	C.free(unsafe.Pointer(name_))
	return retval
}

func (self *PulseMainLoop) Start() int {
	return int(C.pa_threaded_mainloop_start(self.pa))
}

func (self *PulseMainLoop) Dispose() {
	C.pa_threaded_mainloop_free(self.pa)
}

func NewPulseMainLoop() *PulseMainLoop {
	pa := C.pa_threaded_mainloop_new()
	if pa == nil {
		return nil
	}
	return &PulseMainLoop{pa: pa}
}
