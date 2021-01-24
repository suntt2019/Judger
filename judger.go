package judger

/*
#cgo pkg-config: libseccomp
#include "seccomp_rules.h"
#include "stdlib.h"
*/
import "C"
import "unsafe"

const (
	ArgsMaxNumber = 256
	EnvMaxNumber  = 256
)

/*
Config is a struct used to record the running configuration.

MaxCPUTime (ms): max cpu time this process can cost, -1 for unlimited
MaxRealTime (ms): max time this process can run, -1 for unlimited
MaxMemory (byte): max size of the process' virtual memory (address space), -1 for unlimited
MaxStack (byte): max size of the process' stack size
MaxProcessNumber: max number of processes that can be created for the real user id of the calling process, -1 for unlimited
MaxOutputSize (byte): max size of data this process can output to stdout, stderr and file, -1 for unlimited
MemoryLimitCheckOnly: if this value equals 0, we will only check memory usage number, because setrlimit(maxrss) will cause some crash issues
ExePath: path of file to run
InputPath: redirect content of this file to process's stdin
OutputPath: redirect process's stdout to this file
ErrorPath: redirect process's stderr to this file
Args (string array terminated by NULL): arguments to run this process
Env (string array terminated by NULL): environment variables this process can get
LogPath: judger log path
SeccompRuleName(string or NULL): seccomp rules used to limit process system calls. Name is used to call corresponding functions.
Uid: user to run this process
Gid: user group this process belongs to
*/
type Config struct {
	MaxCPUTime           int
	MaxRealTime          int
	MaxMemory            int32
	MaxStack             int32
	MaxProcessNumber     int
	MaxOutputSize        int32
	MemoryLimitCheckOnly int
	ExePath              string
	InputPath            string
	OutputPath           string
	ErrorPath            string
	Args                 []string
	Env                  []string
	LogPath              string
	SeccompRuleName      string
	Uid                  uint32
	Gid                  uint32
}

/*
Result is a struct used to record the running result.

CPUTime: cpu time the process has used
RealTime: actual running time of the process
Memory: max value of memory used by the process
Signal: signal number
ExitCode: process's exit code
Result: judger result.
SUCCESS = 0
CPU_TIME_LIMIT_EXCEEDED=1
REAL_TIME_LIMIT_EXCEEDED=2
MEMORY_LIMIT_EXCEEDED=3
RUNTIME_ERROR=4
SYSTEM_ERROR=5
Error: args validation error or judger internal error.
SUCCESS = 0
INVALID_CONFIG = -1
FORK_FAILED = -2
PTHREAD_FAILED = -3
WAIT_FAILED = -4
ROOT_REQUIRED = -5
LOAD_SECCOMP_FAILED = -6
SETRLIMIT_FAILED = -7
DUP2_FAILED = -8
SETUID_FAILED = -9
EXECVE_FAILED = -10
SPJ_ERROR = -11
*/
type Result struct {
	CPUTime  int
	RealTime int
	Memory   int32
	Signal   int
	ExitCode int
	Result   int
	Error    int
}

func (c Config) convertToCStruct() (cc C.struct_config) {
	cc.max_cpu_time = C.int(c.MaxCPUTime)
	cc.max_real_time = C.int(c.MaxRealTime)
	cc.max_memory = C.long(c.MaxMemory)
	cc.max_stack = C.long(c.MaxStack)
	cc.max_process_number = C.int(c.MaxProcessNumber)
	cc.max_output_size = C.long(c.MaxOutputSize)
	cc.memory_limit_check_only = C.int(c.MemoryLimitCheckOnly)
	cc.exe_path = C.CString(c.ExePath)
	cc.input_path = C.CString(c.InputPath)
	cc.output_path = C.CString(c.OutputPath)
	cc.error_path = C.CString(c.ErrorPath)
	for i := 0; i < len(c.Args) && i < ArgsMaxNumber-1; i++ {
		cc.args[i] = C.CString(c.Args[i])
	}
	for i := 0; i < len(c.Env) && i < EnvMaxNumber-1; i++ {
		cc.env[i] = C.CString(c.Env[i])
	}
	cc.log_path = C.CString(c.LogPath)
	cc.seccomp_rule_name = C.CString(c.SeccompRuleName)
	cc.uid = C.uint(c.Uid)
	cc.gid = C.uint(c.Gid)
	return
}

func (r *Result) convertFromCStruct(cr C.struct_result) {
	r.CPUTime = int(cr.cpu_time)
	r.RealTime = int(cr.real_time)
	r.Memory = int32(cr.memory)
	r.Signal = int(cr.signal)
	r.ExitCode = int(cr.exit_code)
	r.Result = int(cr.result)
	r.Error = int(cr.error)
}

// Run runs the program in the sandbox according to the config and returns the result.
func Run(config Config) (result Result) {
	var cResult C.struct_result
	cConfig := config.convertToCStruct()
	C.run(&cConfig, &cResult)
	result.convertFromCStruct(cResult)
	C.free(unsafe.Pointer(cConfig.exe_path))
	C.free(unsafe.Pointer(cConfig.input_path))
	C.free(unsafe.Pointer(cConfig.output_path))
	C.free(unsafe.Pointer(cConfig.error_path))
	C.free(unsafe.Pointer(cConfig.log_path))
	C.free(unsafe.Pointer(cConfig.seccomp_rule_name))
	for i := range cConfig.args {
		if i == 0 {
			break
		}
		C.free(unsafe.Pointer(i))
	}
	for i := range cConfig.env {
		if i == 0 {
			break
		}
		C.free(unsafe.Pointer(i))
	}
	return
}
