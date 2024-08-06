# waserver - WEB Application Server
![waserver builder](https://github.com/midstar/waserver/actions/workflows/build.yml/badge.svg)

waserver is a generic WEB Application Server. It consists of one executable 
and no external dependencies.

All majort platforms supported such as Windows, Linux (x86 and ARM)
and Mac.

No installation required, just download the suitable executable 
[here on GitHub](https://github.com/midstar/waserver/releases). Or build
the software yourself according to the instructions below.

## Usage

Copy the waserver to a directory of your choice and add it to your PATH
environment variable.

    Usage: ./waserver [options] [<apppath>] [<datapath>]

    <apppath> is the directory where your web files are located.
    Default is app directory.

    <datapath> is the directory where data (JSON) is stored.
    Default is data directory.

    Supported options:
    -c string
            TLS certificate file (default "cert.pem")
    -d    Enable debugging logs
    -k string
            TLS key file (default "key.pem")
    -p int
            Network port to listen to (default 8080)
    -s    Use secure connection (TLS/HTTPS)
    -v    Display version

The WEB applications (i.e. html, js, css files etc.) are put in the
directory set as apppath.

Open a WEB browser and enter:

    <ipaddress>:<port>

To access the applications.

OpenSSL can be used to generate the public and private key required for TLS/HTTPS:

    openssl genrsa -out key.pem 2048
    openssl req -new -x509 -sha256 -key key.pem -out cert.pem -days 3650

## Installing applications

The applications are ordinary WEB applications utilizing the waserver REST API.

Usually the application consists of an index.html file possible other resources
(css, js, images etc.).

The applications needs to be put in following directory:

    <apppath>/<applicationname>/

To get a nice logo image in the waserver start page you need to add an image 
called logo.ico inside the applicationname directory.

## REST API

To build an application following REST API is exposed by waserver. wasserver
reads or stores javascript objects as ordinary text files under the data
directory (default data).

**&lt;addr&gt;** is the IP address and port of waserver.

**&lt;directories&gt;** are optional and any number of directories can be
added such as:

    192.168.1.104:8080/data/this/is/a/deep/dir/structure/myobj

Directories are created by wasserver if they do not exist.

**NOTE!** waserver has no authentification or any other security protection.
Applications are not isolated from each other and might overwrite or delete each
others data. 

### GET &lt;addr&gt;/data/&lt;directories&gt;/&lt;objname&gt;

Get javascript object with name &lt;objname&gt;. 

### POST &lt;addr&gt;/data/&lt;directories&gt;/&lt;objname&gt;

Write (or overwrite) javascript object with name &lt;objname&gt;.

### DELETE &lt;addr&gt;/data/&lt;directories&gt;/&lt;objname&gt;

Delete javascript object with name &lt;objname&gt;.  

### GET &lt;addr&gt;/data/&lt;directories&gt;/&lt;dirname&gt;/

**Note that dirname needs to end with /**

Get a javascript object including all objects inside &lt;dirname&gt;/.
The entries are named after the object names and values are the contents. 

For example:

    {
        "obj1" : <obj1 contents>,
        "obj2" : <obj2 contents>
    }

### GET &lt;addr&gt;/data/&lt;directories&gt;/&lt;dirname&gt;/?ls=true;

**Note that dirname needs to end with /**

List all files inside directory. Returns following javascript object:

    {
      "files" : ["file1", "file2", ...]
      "dirs" : ["dir1", "dir2", ...]
    }

### DELETE &lt;addr&gt;/data/&lt;directories&gt;/&lt;dirname&gt;/

**Note that dirname needs to end with /**

Delete directory with name &lt;dirname&gt;/.

## Build from source (any platform)

To build from source on any platform you need to:

* Install Golang 
* Set the GOPATH environment variable

Then run:

    go get github.com/midstar/waserver
    go install github.com/midstar/waserver


## Author and license

This application is written by Joel Midstjärna and is licensed under the MIT License.