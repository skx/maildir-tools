# maildir-utils

In the past I've written a couple of console-based email clients, and I always found the user-interface the hardest part.

This repository contains a simple proof of concept for a different approach towards email clients:

* Instead of a monolithic "mail-client" why not compose one from pieces?

> This is basically the motivation and history behind MH-E.


# Usage

There are currently two sub-commands:

* `maildir-utils maildirs`
  * This lists all your maildir folders, recursively.
* `maildir-utils messages`
  * This lists the messages inside a folder.
  * Handling the output in a flexible fashion.


## `maildir-utils maildirs`

This will output a list of maildir directories, by default showing the complete path.  You can add `-short` to view only the name of the maildir itself

```
$ maildir-utils maildirs | grep debian-packages
/home/skx/Maildir/debian-packages
/home/skx/Maildir/debian-packages-Pkg-javascript-devel
/home/skx/Maildir/debian-packages-aatv
/home/skx/Maildir/debian-packages-abiword
/home/skx/Maildir/debian-packages-amaya
/home/skx/Maildir/debian-packages-anon-proxy
/home/skx/Maildir/debian-packages-apache
/home/skx/Maildir/debian-packages-apache2
/home/skx/Maildir/debian-packages-apachetop
/home/skx/Maildir/debian-packages-apt
/home/skx/Maildir/debian-packages-apt-listchanges
```

## `maildir-utils messages`

This is the star of the show!  It allows you to list the messages contained
within a maildir folder - with a flexible formatting system for output

```
$ maildir-utils messages --format '${flags} ${from} ${subject}' debian-packages-ttylog
S Steve Kemp <skx@debian.org> Bug#553945: Acknowledgement (ttylog: Doesn't test length of device.)

```

Here `${flags}` was replaced by the message flags (`S` in this case), `${from}` with the message sender, etc.

(You can also use `${file}` to refer to the filename of the message, and other header-values as you would expect.)


# TODO

I could decode to allow the display of a simple message:

```
$ maildir-utils message /path/to/foo.msg:2,S
```

After that a simple command to set-flags (i.e. mark as read/replied).

And exec vi/emacs to compose/reply.
