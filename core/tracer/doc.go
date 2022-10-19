/*
Package tracer helps with the propagation of the tracer span through context
in the system. It does this for operations contained to single node, as well as
across nodes, by injecting special headers.

To use the tracer package, a Tracer instance must be created, which contains
functions for starting new span contexts, injecting them in other data, and
extracting the active span them from the context.

To use the tracer package a Tracer instance must be created:

	tracer, tracerCloser, err := tracer.NewTracer(&tracer.Options{
		Enabled:     true,
		Endpoint:    "127.0.0.1:6831",
		ServiceName: "mop",
	})
	if err != nil {
		// handle error
	}
	defer tracerCloser.Close()
	// ...

The tracer instance contains functions for starting new span contexts, injecting
them in other data, and extracting the active span them from the context:

	span, _, mctx := tracer.StartSpanFromContext(mctx, "operation-name", nil)

Once the operation is finished, the open span should be finished:

	span.Finish()

The tracer package also provides a function for creating a logger which will
inject a "traceID" field entry to the log line, which helps in finding out which
log lines belong to a specific trace.

To create a logger with trace just wrap an existing logger:

	logger := tracer.NewLoggerWithTraceID(mctx, s.logger)
	// ...
	logger.Info("some message")

Which will result in following log line (if the context contains tracer
information):

	time="2015-09-07T08:48:33Z" level=info msg="some message" traceID=ed65818cc1d30c
*/
package tracer
