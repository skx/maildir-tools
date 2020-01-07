# maildir-utils

In the past I've written a couple of console-based email clients, and I always found the user-interface the hardest part.

This repository contains a simple proof of concept for a different approach towards email clients:

* Instead of a monolithic "mail-client" why not compose one from pieces?

> This is basically the motivation and history behind MH-E.

I can imagine a UI which is stateful:

* In one state it just runs a shell-command to list folders, and lets you move a cursor up and down.
* In another state it might just run a shell-command to list messages in the folder you've chosen
  * And allow you to move the cursor up and down.
* In the final state it might just run a shell-command to display a message.
  * And allow you to hit keys to mark read, reply, etc.

This should be almost trivial to write.  Right?  The hardest part would be handling the sorting of messages into threads, etc.



# Usage

There are currently two sub-commands:

* `maildir-utils maildirs`
  * This lists all your maildir folders, recursively.
* `maildir-utils messages`
  * This lists the messages inside a folder.
  * Handling the output in a flexible fashion.

Both of these default to looking in `~/Maildir` but the `-prefix /path/to/root` will let you change the directory.  Maildirectories are handled recursively, and things are pretty fast but I guess local SSDs help with that.  For everything else there is always the option to cache things.


# Sub-Commands

This is a simple proof of concept, it might become more useful, it might become abandoned.

Currently this project will build a single monolithic binary with a couple of sub-commands, run with no arguments to see usage information.


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

You can specify either the short-path to the Maildir, beneath the root directory, or the complete path `/home/skx/Maildir/people-foo`, depending upon your preference.



# TODO

I could decode and display a single message, like so:

```
$ maildir-utils message /path/to/foo.msg:2,S
```

Of course how to display attachments is a harder question.  If we had
 a real UI we'd just present a list.  Decoding a message is more useful than
 exec'ing `less`, partly due to MIME and partly due to consistency.

After that a simple command to set-flags (i.e. mark as read/replied).

And exec vi/emacs to compose/reply.
