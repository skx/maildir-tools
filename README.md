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

There are currently three sub-commands:

* `maildir-utils maildirs`
  * This lists all your maildir folders, recursively.
* `maildir-utils messages`
  * This lists the messages inside a folder.
  * Handling the output in a flexible fashion.
* `maildir-utils ui`
  * Toy user-interface that proves we could make something of ourselves.

Both of these default to looking in `~/Maildir` but the `-prefix /path/to/root` will let you change the directory.  Maildirectories are handled recursively, and things are pretty fast but I guess local SSDs help with that.  For everything else there is always the option to cache things.


# Sub-Commands

This is a simple proof of concept, it might become more useful, it might become abandoned.

Currently this project will build a single monolithic binary with a couple of sub-commands, run with no arguments to see usage information.


## `maildir-utils maildirs`

This will output a list of maildir directories, by default showing the complete path of each maildir which was found.

```
$ maildir-utils maildirs --format '${name}' | grep debian-packages
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

Flags can be used to refine the output, for example:

* `-short` to view only the name of the maildir itself
  * e.g. "debian-packages", "debian-packages-abiword", etc.
* `-format '${unread} ${total} ${name}'`
  * To specify what is output.
* `-unread`
  * Only show maildirs containing unread messages.


## `maildir-utils messages`

This is the star of the show!  It allows you to list the messages contained
within a maildir folder - with a flexible formatting system for output

```
$ maildir-utils messages --format '${flags} ${from} ${subject}' debian-packages-ttylog
S Steve Kemp <skx@debian.org> Bug#553945: Acknowledgement (ttylog: Doesn't test length of device.)

```

Here `${flags}` was replaced by the message flags (`S` in this case), `${from}` with the message sender, etc.

(You can also use `${file}` to refer to the filename of the message, and other header-values as you would expect.  The other magical values are "`${index}/${total}`" which show the current/total number of entries.)

You can specify either the short-path to the Maildir, beneath the root directory, or the complete path `/home/skx/Maildir/people-foo`, depending upon your preference.


## `maildir-utils ui`

This is an __extremely__ minimal UI, which shows a list of maildirs.

It works by executing itself, which is suboptimal.

You can view/scroll the message lists.  `vi` keys work.  As do others.



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


## Plan

* [x] Allow listing maildirs with a format string
  * "`${name} ${unread} ${total}`"
* [x] Fix `maildir-utils maildirs -unread` to work.
* [ ] Fix RFC2047-decoding of message-subjects.
* [ ] Move the prefix-handling to a common-library.
* [ ] Move the formatting of a message-list to a common-library.
* [ ] Consider a caching-plan
* [ ] Consider how threading would be handled, or even sorting of messages.
* [ ] Consider a message-view
* [x] Sketch out a console UI to prove it is even worthwhile, possible.
  * [x] Start with maildir-view
  * [ ] Then allow message-list-view
  * [ ] Then message-view
  * [ ] Modal/Stateful
  * [ ] Should essentially `exec $self $mode`
    * [ ] Cache the output to RAM?  File?
    * [ ] When to refresh?
    * [ ] Display the output.  But modify it
       * e.g. We can change "/home/skx/Maildir/xxx" to "XXX" in the display
       * But we want the full-path to know what to enter when the user chooses the directory
       * Similarly when viewing a message-list we'll need ${file} to know what to view, but we probably don't want to display that on-screen.


Displaying a message will probably be done via a text/template, like so:

```
TO: ${to}
From:${from}
Subject: ${subject}
Date: ${date}
Flags: ${flags}

${Body:Text}
${Attachment-Names}
```

But we'll need to allow the user to specify their own, and allow arbitrary
header values to be shown.  Or even toggled.
