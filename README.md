# ssd

## Name

*ssd* - Service SystemD, remotely interact with systemd.

## Description

*ssd* consists out of a server `ssdd` and a client `ssdc` that work to together to interact with the
`systemd` running on the system where `ssdd` also runs. The daemon exposes a REST api, the client
uses.

When `ssdd` starts up it can potentially load unit files from a directory.

## Syntax

~~~
ssdd [-auth] [-users FILE] [-load DIR] [-p PORT]
~~~

* `-auth`, authenticate the request and perform authorizaton check in ssdd
* `-users FILE`, used **FILE** to lookup the users for authorization.
* `-p PORT`, bind to port **PORT**
* `-load DIR`, look in **DIR** for unit files to load (.service, .path)

`ssdd` will call into the user systemd that runs under the same user as itself. The file used for
user authorization is simplistic, it is just one user name per line.

The default port is 9999. Metrics are served from the same port on the /metrics path. The same port
also provides a /health handler that return "200 OK".

It's assumed `ssdd` is ran via systemd as well, so all interactions it provides can be applied to
itself.

The client the syntax is as follows:
~~~
ssdc ADDRESS OPERATION [SERVICE]
~~~
* where **ADDRESS** is the endpoint where `ssdd` runs
* **OPERATION** is the systemd operation you want to perform, see below
* **SERVICE** is the service you are operating on, may include the extension

There is no provisioning made to operate on multiple addresses and/or services, as to
reflect any remote state into the exit status of `ssdc` itself. Chaining of services should be
handled by the systemd unit files and/or in a script calling `ssdc`.

The following **OPERATION**s are available:

* `list` - list the loaded unit files, the **SERVICE** is not used in this case and should be
  omitted. Only .service and .path unit files are shown, This calls `list-units`
* `cat` - show the contents of the unit file for **SERVICE**
* `start`, `status`, `stop`, `reload`, and `restart`, will call this respective command
* `logs`, show the logs of **SERVICE**, assuming they are in journald.

## Protocol

The protocol is HTTP/HTTPS. Cookies are used for authentication and REST structure mimics the `ssdc`
commandline:

~~~
/s/OPERATION[/SERVICE]?OPTIONS
~~~

Where `s` stands for systemd, and **OPERATION** and **SERVICE** are the same as above. **OPTIONS**
are **OPERATION** specific options that some commmand allow, first and foremost the `logs` one.
There is one mandatory option which is the user issuing the operation. Thus the following options
are defined:

* ...

### Authentication and Authorizaton

A cookie should be included in the request, which conveys 'I have been authenticated'. This is
implemented in a vendor specific way.

For authorization the user is looked up a table (=file). If there user is present there the
operations is allowed, otherwise it is denied.

If the file is empty, the system will fail open and all user are allowed to execute operations.

## Metrics

The following metrics are exposed:

* `ssdd_request_total`, a counter of all incoming requests
* `ssdd_errors_total`, a counter of all requests leading to an error

## Notes

Instead of connecting to the systemd DBUS, `ssdd` will simply call `systemd`, this is needed for
getting the logs anyway, so it's extended to the entire binary. This also prevents `sssd` from
needing to use a C library, which means it can't be a statically compiled Go binary.

## Bugs

The -load option has not been implemented. The user file might be a bad idea.
