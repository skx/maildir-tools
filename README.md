
This repository contains a simple golang utility which can be used for two purposes:

* To perform simple script operations against Maildir hierarchies.
* To provide a simple, console-based, email-client.
  * Albeit a basic one that only allows reading/viewing maildirs/messages.

There is a [demo of mail-client UI](https://asciinema.org/a/FXjgOsnwjVu0lB5znx8EwRVWF), but the focus at the moment is upon improving the scripting facilities.


# Contents

* [maildir-tools](#maildir-tools)
* [Installation](#installation)
* [Usage](#usage)
* [Sub-Commands](#sub-commands)
  * [maildir-tools maildirs](#maildir-tools-maildirs)
  * [maildir-tools messages](#maildir-tools-messages)
  * [maildir-tools message](#maildir-tools-message)
  * [maildir-tools ui](#maildir-tools-ui)


# maildir-tools

In the past I've written a couple of console-based email clients, and I always found the user-interface the hardest part.

This repository contains a simple proof of concept for a different approach towards email clients:

* Instead of a monolithic "mail-client" why not compose one from pieces?
  * This is basically the motivation and history behind [MH](https://en.wikipedia.org/wiki/MH_Message_Handling_System)

I can imagine a UI which is nothing more than a bunch of shell-scripts, perhaps using `dialog` to drive them:

* In one state it just runs a shell-command to list folders, and lets you move a cursor up and down to enter one.
* In another state it might just run a shell-command to list messages in the folder you've chosen.
  * Again you may move the cursor up and down, and select a message.
* In the final state a shell-command might be executed to display a message.
  * And allow you to hit keys to mark read, reply, etc.

This should be almost trivial to write.  Right?  The hardest part would be handling the sorting of messages into threads, etc.

The driving force behind this repository is currently the ability to provide the primitives that a simple UI would need - and as a side-effect provide a useful tool for working with Maildir hierarchies.

It might be that, over time, a console-based mail-client is built, but that is a secondary focus.  I personally have thousands of email messages in a deeply nested Maildir hierarchy and even being able to dump, script, and view messages in a flexible fashion is useful.  Of course I have written email clients (plural) in the past so it is tempting to try again, but I can appreciate that gaining users is hard and the job is always bigger than expected.


# Installation

To install from source you can use the standard golang-approach:

```
$ go get github.com/skx/maildir-tools/cmd/maildir-tools
```

(If you have the source-repository cloned locally run `cd cmd/maildir-tools && go install .`)

In the future, after we've had our first release, you will be able to download binaries instead.



# Usage

There are currently several sub-commands available:

* `maildir-tools maildirs`
  * This command lists all your maildir folders, recursively.
* `maildir-tools messages $folder1 $folder2 .. $folderN`
  * This lists the messages inside a folder.
  * Handling the output in a flexible fashion.
* `maildir-tools message $file`
  * This formats and displays a single message.
  * If a `text/plain` part is available then that is displayed.
     * Otherwise any available `text/html` part is used.
        *  If neither are available nothing is shown.
* `maildir-tools ui`
  * Toy user-interface that proves we could make something of ourselves.

With the ability to view folders, message-lists, and a single message we almost have enough to be a functional mail-client.  Albeit one within which you cannot compose, delete, or reply to a message.

Most of the sub-commands default to looking in `~/Maildir` but the `-prefix /path/to/root` will let you change the directory.  Maildirectories are handled recursively, and things are pretty fast but I guess local SSDs help with that.  For everything else there is always the option to cache things.



# Sub-Commands

Currently this project will build a single monolithic binary with a couple of sub-commands, run it with no arguments to see usage information.


## `maildir-tools maildirs`

This will output a list of maildir directories, by default showing the complete path of each maildir which was found.

```
$ maildir-tools maildirs --format '${name}' | grep debian-packages
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

* `-format '${unread} ${total} ${name}'`
  * To specify what is output.
* `-unread`
  * Only show maildirs containing unread messages.


## `maildir-tools messages $folder1 $folder2 .. $folderN`

This sub-command allows you to list the messages contained within a maildir folder with a flexible formatting system for output

```
$ maildir-tools messages --format '[${4flags}] ${from} ${subject}' debian-packages-ttylog
[    S] Steve Kemp <skx@debian.org> Bug#553945: Acknowledgement (ttylog: Doesn't test length of device.)

```

Here `${4flags}` was replaced by the message flags (`S` in this case), and that was padded to be four bytes long, `${from}` with the message sender, etc.

(You can also use `${file}` to refer to the filename of the message, and other header-values as you would expect.  The other magical values are "`${index}/${total}`" which show the current/total number of entries.)

You can specify either the short-path to the Maildir, beneath the root directory, or the complete path `/home/skx/Maildir/people-foo`, depending upon your preference.

* I recently switched to a new MIME library, it is slow.
* Takes 20-25 seconds to open a maildir with 40,000 messages.
  * This means I need caching, or need a faster MIME-aware mail-library.
  * Pointers welcome.  Patches even more welcome.


## `maildir-tools ui`

This is an __extremely__ minimal UI, which allows you to navigate:

* When launched you'll be presented with a list of Maildir folders.
* Scrolling up/down works as expected.
* Pressing return will open the maildir, replacing the view with a list of messages.
* Once again scrolling should be responsive and reliable.
* Finally you can view a message by hitting return within the list.

In each case you can return to the previous mode/view via `q`, or quit globally via `Q`.  When you're viewing a single message "`J`" and "`K`" move backwards/forwards by one message.

`vi` keys work, as do HOME, END, PAGE UP|DOWN, etc.


## `maildir-tools message`

This sub-command outputs a reasonably attractive rendering of a single message.
