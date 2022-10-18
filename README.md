# Platform Developer Test

## Task

### Simple implementation of file server based on http.FileServer handle ( https://pkg.go.dev/net/http#example-FileServer )

The server instance is running on top of simple file folder which doesn’t have nested subfolders.

Please implement client which downloads files using this server.

You should download a file containing char 'A' on earlier position than other files.

In case several files have the 'A' char on the same the earliest position you should download all of them.

Each TCP connection is limited by speed. The total bandwidth is unlimited.

You can use any disk space for temporary files.

The goal is to minimize execution time and data size to be transferred.

Example

If the folder contains the following files on server:

'file1' with contents: "---A-------"

'file2' with contents: "--A------"  

'file3' with contents: "------------"

'file4' with contents: "==A=========="

then 'file2' and 'file4' should be downloaded

## How to execute a program

1. Run a httpserver locally which hosts all the files. Imagine such server running at 

    http://localhost:8080

2. Clone the git repo
3. Launch the program using:

        go run main.go

4. To download all the files which matches criteria, run the curl comaand or use postman.

    Method : POST

    URL : http://localhost:8000/files
    
    Body :
    {
    "remote_file_server_url" : "http://localhost:8080",
    "lookup_character" : "C"
    }

Note: The program is reading and downloading files concurrently and can download multiple GB's of file.

Improvements:
1) Populate map only if you find a file with character at smaller position.
2) Dont read the other entire file, if you have already found a file previously having character at earlier position.
3) If a character is found in 3rd chunk for ex, then need to adjust logic to correctly calculate index of character.
4) Make use of tmp file to store the read data, this will decrease the bandwidth used.
