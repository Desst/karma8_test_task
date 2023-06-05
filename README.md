# karma8_test_task

## karma_task
Storage service with 6 server nodes(disk storage) at startup

'karma_task' binary - starts a REST server at 127.0.0.1:8080 with 2 simple handles:
    - POST /api/v1/file?filename=<file_name>
    - GET /api/v1/file?filename=<file_name>

After 1 minute a 7th node is added to the service.

## test_karma

'test_karma' - test binary which generates random 10MBytes buffer and uploads
it as a file to 'karma_task' running service. Then downloads the uploaded file and 
compares it to generated buffer.

After 1 minute when the 7th node is added, uploads the generated buffer again as another file.
Then downloads and compares it to the first file.




### All constants are used for demonstration and test purposes only.
