package utils

import (
	"bytes"
	"runtime"
	"strconv"
)

func TakeStacktrace(skip int) string {
	buffer := bytes.NewBuffer(make([]byte, 0, 1000))
	programCounters := make([]uintptr, 10000)

	var numFrames int
	for {
		// Skip the call to runtime.Callers and takeStacktrace so that the
		// program counters start at the caller of takeStacktrace.
		numFrames = runtime.Callers(skip+2, programCounters)
		if numFrames < len(programCounters) {
			break
		}
		// Don't put the too-short counter slice back into the pool; this lets
		// the pool adjust if we consistently take deep stacktraces.
		programCounters = make([]uintptr, len(programCounters) * 2)
	}

	i := 0
	frames := runtime.CallersFrames(programCounters[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if i != 0 {
			buffer.WriteByte('\n')
		}
		i++
		buffer.WriteString(frame.Function)
		buffer.WriteByte('\n')
		buffer.WriteByte('\t')
		buffer.WriteString(frame.File)
		buffer.WriteByte(':')
		buffer.WriteString(strconv.Itoa(frame.Line))
	}

	return buffer.String()
}
