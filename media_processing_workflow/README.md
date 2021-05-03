## Media Process Workflow
The Media Processing workflow shows an end-to-end workflow demonstration of downloading, encoding, and combining video files.


## Code Organization
For this example, we have 3 main parts to start to get the example working: 
- a simulated, simple vendor API in the `vendor_api` directory
- the worker that hosts our workflow and activities in the `worker` directory
- the starter program (located in the `starter` directory) is a convenience file to trigger our workflow.

The `vendor_api` simulates an external API that indicates status. We include a `media_success_ratio` to simulate 
the fraction of time the API returns success or failure. 


## Prerequisites 
1. Ensure that you have the temporal service running as specified in the quick start of
the [temporal docs](https://docs.temporal.io/docs/server-quick-install/).

2. For this example, we'll need to have ffmpeg installed. See instructions in the [official website](https://ffmpeg.org/download.html). 


## How to run the workflow

1. Start the vendor api by going to the `vendor_api` directory and starting the server:
```
go run *.go
```

1. Start the internal api by going to the `internal_api` directory and starting the internal server:
```
go run *.go
```
The uploaded files will be stored within the `uploadedfiles` directory in the `internal_api` directory.


3. Start the worker by going to the `worker` directory and starting the worker:
```
go run *.go
```
Note: It's possible to instatiate multiple workers by repeatedly running the command above.

4. Trigger the workflow by going to the `starter` directory and running the following command:
```
go run *.go
``` 