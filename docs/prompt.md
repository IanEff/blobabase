Before your interview, write a program that runs a server that is accessible on `http://localhost:4000/`. When your server receives a request on `http://localhost:4000/set?somekey=somevalue` it should store the passed key and value in memory. When it receives a request on `http://localhost:4000/get?key=somekey` it should return the value stored at `somekey`.

During your interview, you'll pair on improving your server. For example, you might decide to save the data to a file; you could start with simply appending each write to the file, and work on making it more efficient if you have time.
