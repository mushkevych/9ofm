package commander

import (
	"testing"
)

type defaultResolver struct{}

// func showEvents(events []testEvent) {
// 	for _, e := range events {
// 		fmt.Printf("{stdout:\"%s\", stderr:\"%s\", errorOnExit: %v, errMessage: \"%s\"},\n",
// 			strings.Replace(vtclean.Clean(e.stdout, false), "\n", "\\n", -1),
// 			strings.Replace(vtclean.Clean(e.stderr, false), "\n", "\\n", -1),
// 			e.errorOnExit,
// 			e.errMessage)
// 	}
// }

type testEvent struct {
	stdout      string
	stderr      string
	errMessage  string
	errorOnExit bool
}

func newTestEvent(e event) testEvent {
	var errMsg string
	if e.err != nil {
		errMsg = e.err.Error()
	}
	return testEvent{
		stdout:      e.stdout,
		stderr:      e.stderr,
		errMessage:  errMsg,
		errorOnExit: e.errorOnExit,
	}
}

func TestRun(t *testing.T) {
	// table := map[string]struct {
	// 	resolver image.Resolver
	// 	options  Options
	// 	events   []testEvent
	// }{
	// 	"fetch-case": {
	// 		resolver: &defaultResolver{},
	// 		options: Options{
	// 			Image:      "dive-example",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			BuildArgs:  nil,
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Image Source: docker://dive-example", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Fetching image... (this can take a while for large images)", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Analyzing image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Building cache...", stderr: "", errorOnExit: false, errMessage: ""},
	// 		},
	// 	},
	// 	"fetch-with-no-build-options-case": {
	// 		resolver: &defaultResolver{},
	// 		options: Options{
	// 			Image:      "dive-example",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			// note: empty slice is passed
	// 			BuildArgs: []string{},
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Image Source: docker://dive-example", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Fetching image... (this can take a while for large images)", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Analyzing image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Building cache...", stderr: "", errorOnExit: false, errMessage: ""},
	// 		},
	// 	},
	// 	"build-case": {
	// 		resolver: &defaultResolver{},
	// 		options: Options{
	// 			Image:      "dive-example",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			BuildArgs:  []string{"an-option"},
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Building image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Analyzing image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Building cache...", stderr: "", errorOnExit: false, errMessage: ""},
	// 		},
	// 	},
	// 	"failed-fetch": {
	// 		resolver: &failedFetchResolver{},
	// 		options: Options{
	// 			Image:      "dive-example",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			BuildArgs:  nil,
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Image Source: docker://dive-example", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Fetching image... (this can take a while for large images)", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "", stderr: "cannot fetch image", errorOnExit: true, errMessage: "some fetch failure"},
	// 		},
	// 	},
	// 	"failed-build": {
	// 		resolver: &failedBuildResolver{},
	// 		options: Options{
	// 			Image:      "doesn't-matter",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			BuildArgs:  []string{"an-option"},
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Building image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "", stderr: "cannot build image", errorOnExit: true, errMessage: "some build failure"},
	// 		},
	// 	},
	// 	"ci-go-case": {
	// 		resolver: &defaultResolver{},
	// 		options: Options{
	// 			Image:      "doesn't-matter",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			BuildArgs:  []string{"an-option"},
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Building image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Analyzing image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "  efficiency: 98.4421 %", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "  wastedBytes: 32025 bytes (32 kB)", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "  userWastedPercent: 48.3491 %", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Inefficient Files:\nCount  Wasted Space  File Path\n    2         13 kB  /root/saved.txt\n    2         13 kB  /root/example/somefile1.txt\n    2        6.4 kB  /root/example/somefile3.txt\nResults:\n  FAIL: highestUserWastedPercent: too many bytes wasted, relative to the user bytes added (%-user-wasted-bytes=0.4834911001404049 > threshold=0.1)\n  FAIL: highestWastedBytes: too many bytes wasted (wasted-bytes=32025 > threshold=1000)\n  PASS: lowestEfficiency\nResult:FAIL [Total:3] [Passed:1] [Failed:2] [Warn:0] [Skipped:0]\n", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "", stderr: "", errorOnExit: true, errMessage: ""},
	// 		},
	// 	},
	// 	"empty-ci-config-case": {
	// 		resolver: &defaultResolver{},
	// 		options: Options{
	// 			Image:      "doesn't-matter",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "",
	// 			BuildArgs:  []string{"an-option"},
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Building image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Analyzing image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "  efficiency: 98.4421 %", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "  wastedBytes: 32025 bytes (32 kB)", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "  userWastedPercent: 48.3491 %", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Inefficient Files:\nCount  Wasted Space  File Path\nNone\nResults:\n  MISCONFIGURED: highestUserWastedPercent: invalid config value (''): strconv.ParseFloat: parsing \"\": invalid syntax\n  MISCONFIGURED: highestWastedBytes: invalid config value (''): strconv.ParseFloat: parsing \"\": invalid syntax\n  MISCONFIGURED: lowestEfficiency: invalid config value (''): strconv.ParseFloat: parsing \"\": invalid syntax\nCI Misconfigured\n", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "", stderr: "", errorOnExit: true, errMessage: ""},
	// 		},
	// 	},
	// 	"export-go-case": {
	// 		resolver: &defaultResolver{},
	// 		options: Options{
	// 			Image:      "doesn't-matter",
	// 			Source:     dive.SourceDockerEngine,
	// 			ExportFile: "some-file.json",
	// 			BuildArgs:  []string{"an-option"},
	// 		},
	// 		events: []testEvent{
	// 			{stdout: "Building image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Analyzing image...", stderr: "", errorOnExit: false, errMessage: ""},
	// 			{stdout: "Exporting image to 'some-file.json'...", stderr: "", errorOnExit: false, errMessage: ""},
	// 		},
	// 	},
	// }

	// for name, test := range table {
	// 	var ec = make(eventChannel)
	// 	var events = make([]testEvent, 0)
	// 	var filesystem = afero.NewMemMapFs()

	// 	go start(false, test.options, test.resolver, ec, filesystem)

	// 	for event := range ec {
	// 		events = append(events, newTestEvent(event))
	// 	}

	// 	// fmt.Println(name)
	// 	// showEvents(events)
	// 	// fmt.Println()

	// 	if len(test.events) != len(events) {
	// 		t.Fatalf("%s.%s: expected # events='%v', got '%v'", t.Name(), name, len(test.events), len(events))
	// 	}

	// 	for idx, actualEvent := range events {
	// 		expectedEvent := test.events[idx]

	// 		if expectedEvent.errorOnExit != actualEvent.errorOnExit {
	// 			t.Errorf("%s.%s: expected errorOnExit='%v', got '%v'", t.Name(), name, expectedEvent.errorOnExit, actualEvent.errorOnExit)
	// 		}

	// 		actualEventStdoutClean := vtclean.Clean(actualEvent.stdout, false)
	// 		expectedEventStdoutClean := vtclean.Clean(expectedEvent.stdout, false)

	// 		if expectedEventStdoutClean != actualEventStdoutClean {
	// 			t.Errorf("%s.%s: expected stdout='%v', got '%v'", t.Name(), name, expectedEventStdoutClean, actualEventStdoutClean)
	// 		}

	// 		actualEventStderrClean := vtclean.Clean(actualEvent.stderr, false)
	// 		expectedEventStderrClean := vtclean.Clean(expectedEvent.stderr, false)

	// 		if expectedEventStderrClean != actualEventStderrClean {
	// 			t.Errorf("%s.%s: expected stderr='%v', got '%v'", t.Name(), name, expectedEventStderrClean, actualEventStderrClean)
	// 		}

	// 		if expectedEvent.errMessage != actualEvent.errMessage {
	// 			t.Errorf("%s.%s: expected error='%v', got '%v'", t.Name(), name, expectedEvent.errMessage, actualEvent.errMessage)
	// 		}

	// 		if test.options.ExportFile != "" {
	// 			if _, err := filesystem.Stat(test.options.ExportFile); os.IsNotExist(err) {
	// 				t.Errorf("%s.%s: expected export file but did not find one", t.Name(), name)
	// 			}
	// 		}
	// 	}
	// }
}
