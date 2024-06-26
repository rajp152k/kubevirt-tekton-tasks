package parse_test

import (
	"time"

	"github.com/kubevirt/kubevirt-tekton-tasks/modules/execute-in-vm/pkg/utils/parse"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap/zapcore"

	"reflect"
)

var (
	defaultNS       = "default"
	script          = "#!/bin/bash\necho hello world"
	commandArr      = []string{"echo", "-E", "hello", "world"}
	expectedCommand = "echo -E hello world"
)

var _ = Describe("CLIOptions", func() {
	DescribeTable("Init return correct assertion errors", func(expectedErrMessage string, options *parse.CLIOptions) {
		Expect(options.Init().Error()).To(ContainSubstring(expectedErrMessage))
	},
		Entry("no vm", "missing value for vm-name option", &parse.CLIOptions{
			VirtualMachineNamespace: defaultNS,
		}),
		Entry("invalid vm name", "vm-name is not a valid name: a lowercase RFC 1123 subdomain must consist of", &parse.CLIOptions{
			VirtualMachineName:      "no dns 1123",
			VirtualMachineNamespace: defaultNS,
		}),
		Entry("no script or command", "no action was specified: at least one of the following options is required: command|script|stop|delete", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
		}),
		Entry("script and command", "one of command|script options is allowed", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			Command:                 commandArr,
		}),
		Entry("no connection secret", "connection secret should not be empty", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
		}),
		Entry("empty connection secret", "connection secret should not be empty", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			ConnectionSecretName:    "__empty__",
		}),
		Entry("invalid connection secret", "connection secret does not have a valid name: a lowercase RFC 1123 subdomain must consist of", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			ConnectionSecretName:    "secret!",
		}),
		Entry("invalid timeout", "could not parse timeout: time: unknown unit", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			Timeout:                 "1h5q",
			ConnectionSecretName:    "my-secret",
		}),
		Entry("invalid stop", "invalid option stop stahp, only true|false is allowed", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			Stop:                    "stahp",
			ConnectionSecretName:    "my-secret",
		}),
		Entry("invalid delete", "invalid option delete yes, only true|false is allowed", &parse.CLIOptions{
			VirtualMachineName:      "test",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			Delete:                  "yes",
			ConnectionSecretName:    "my-secret",
		}),
	)
	//
	DescribeTable("Parses and returns correct values", func(options *parse.CLIOptions, expectedOptions map[string]interface{}) {
		Expect(options.Init()).Should(Succeed())

		for methodName, expectedValue := range expectedOptions {
			results := reflect.ValueOf(options).MethodByName(methodName).Call([]reflect.Value{})
			Expect(results[0].Interface()).To(Equal(expectedValue))
		}
	},
		Entry("returns valid defaults", &parse.CLIOptions{
			VirtualMachineName:      "vm",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			ConnectionSecretName:    "my-secret",
		}, map[string]interface{}{
			"GetVirtualMachineNamespace": defaultNS,
			"GetScript":                  script,
			"GetDebugLevel":              zapcore.InfoLevel,
			"GetScriptTimeout":           0 * time.Second,
			"ShouldStop":                 false,
			"ShouldDelete":               false,
		}),
		Entry("handles Script cli arguments", &parse.CLIOptions{
			VirtualMachineName:      "vm",
			VirtualMachineNamespace: defaultNS,
			Script:                  script,
			Debug:                   true,
			Timeout:                 "5m10s",
			Stop:                    "true",
			Delete:                  "false",
			ConnectionSecretName:    "my-secret",
		}, map[string]interface{}{
			"GetVirtualMachineNamespace": defaultNS,
			"GetScript":                  script,
			"GetDebugLevel":              zapcore.DebugLevel,
			"GetScriptTimeout":           5*time.Minute + 10*time.Second,
			"ShouldStop":                 true,
			"ShouldDelete":               false,
		}),
		Entry("handles simple Command cli arguments", &parse.CLIOptions{
			VirtualMachineName:      "vm",
			VirtualMachineNamespace: defaultNS,
			Command:                 []string{"ls"},
			ConnectionSecretName:    "my-secret",
		}, map[string]interface{}{
			"GetVirtualMachineNamespace": defaultNS,
			"GetScript":                  "ls",
		}),
		Entry("handles Command cli arguments", &parse.CLIOptions{
			VirtualMachineName:      "vm",
			VirtualMachineNamespace: defaultNS,
			Command:                 commandArr,
			Debug:                   true,
			Timeout:                 "12h5m10s",
			Stop:                    "true",
			Delete:                  "true",
			ConnectionSecretName:    "my-secret",
		}, map[string]interface{}{
			"GetVirtualMachineNamespace": defaultNS,
			"GetScript":                  expectedCommand,
			"GetDebugLevel":              zapcore.DebugLevel,
			"GetScriptTimeout":           12*time.Hour + 5*time.Minute + 10*time.Second,
			"ShouldStop":                 true,
			"ShouldDelete":               true,
		}),
	)

})
