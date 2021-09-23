# ssd

## Name

* ssd* - Service SystemD, remotely interact with systemd.

## Description

*ssd* consists out of a server `ssdd` and a client `ssdc` that work to together to interact with the
`systemd` running on the system where `ssdd` also runs. The daemon exposes a REST api, the client
uses.

When `ssdd` starts up it can potentially load unit files from a directory.

## Syntax

~~~
ssdd [-noauth] [-load DIR] [-p PORT]
~~~

* `-noauth`, disables the authentication check in ssdd
* `-load DIR`, look in **DIR** for unit files to load (.service, .path)
* `-p PORT`, bind to port **PORT**

`ssdd` will call into the user systemd that runs under the same user as itself.

The client the syntax is as follows:
~~~
ssdc ADDRESS OPERATION [SERVICE]
~~~
* where **ADDRESS** is the endpoint where `ssdd` runs
* **OPERATION** is the systemd operation you want to perform, see below
* **SERVICE** is the service you are operating on, may include the extension

There is no provisioning made to operate on multiple **ADDRESSES** and/or **SERVICE**s, as to
reflect any remote state into the exit status of `ssdc` itself. Chaining of services should be
handled by the systemd unit files and/or in a script calling `ssdc`.

The following **OPERATION**s are available:

* `list` - list the loaded unit files, the **SERVICE** is not used in this case and should be
  omitted. Only .service and .path unit files are shown.
* `cat` - show the contents of the unit file for **SERVICE**
* `start`, `stop`, `reload`, `restart`, will call the respective
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

* `user=USER`, where **USER** is an identifier that identifies the user.

### Authentication and Authorizaton

A cookie should be included in the request, which conveys 'I have been authenticated'. This is
implemented in a vendor specific way.

Authorizaton is TDB.

## Notes

Instead of connecting to the systemd DBUS, `ssdd` will simply call `systemd`, this is needed for
getting the logs anyway, so it's extended to the entire binary. This also prevents `sssd` from
needing to use a C library, which means it can't be a statically compiled Go binary.
