
This repository contains a simple golang utility which can be used for two purposes:

* To perform simple scripting operations against Maildir hierarchies.
* To provide a simple console-based email-client.
  * Albeit a basic one that only allows reading/viewing maildirs/messages.
  * (i.e. You cannot reply, compose, delete, or mark a message as read/unread.)

There is a [demo of mail-client UI](https://asciinema.org/a/FXjgOsnwjVu0lB5znx8EwRVWF), but the focus at the moment is upon improving the scripting facilities.


# Contents

* [maildir-tools](#maildir-tools)
  * [Installation](#installation)
* [Scripting Usage](#scripting-usage)
  * [Scripting Usage: Maildir List](#scripting-usage-maildir-list)
  * [Scripting Usage: Message List](#scripting-usage-message-list)
  * [Scripting Usage: Message Display](#scripting-usage-message-display)
* [Console Mail Client](#console-mail-client)




# maildir-tools

In the past I've written a couple of console-based email clients, and I always found the user-interface the hardest part.

This repository contains a proof of concept for a different approach towards an email client: Instead of a monolithic "mail-client" why not compose one from pieces?  (Which is basically the motivation and history behind [MH](https://en.wikipedia.org/wiki/MH_Message_Handling_System).)

I can imagine mail-client which is nothing more than a bunch of shell-scripts, perhaps using `dialog` to glue them together, because most mail-clients are "modal" in intent, if not necessarily in operation:

* We need to view a list of folders.
* We need to view a list of messages.
* We need to view a single message.

The scripting support of `maildir-tools` allows each of those to be handled in a simple, reliable, and flexible fashion.  The driving force behind this repository is providing those primitives, things that a simple UI would need - and as a side-effect provide a useful tool for working with Maildir hierarchies.

It might be that, over time, a console-based mail-client is built, but that is a secondary focus.  I personally have thousands of email messages in a deeply nested Maildir hierarchy and even being able to dump, script, and view messages in a flexible fashion is useful.  Of course I have written email clients (plural!) in the past so it is tempting to try again, but I can appreciate that gaining users is hard and the job is always bigger than expected.




## Installation

To install from source you can use the standard golang-approach:

```
$ go get github.com/skx/maildir-tools/cmd/maildir-tools
```

(If you have the source-repository cloned locally run `cd cmd/maildir-tools && go install .`)

In the future, after we've had our first release, you will be able to download binaries instead.



# Scripting Usage

There are several sub-commands available which are designed to allow you to script access to a maildir-hierarchy, and the messages stored within it.

* `maildir-tools maildirs`
  * This command lists all your maildir folders, recursively.
* `maildir-tools messages $folder1 $folder2 .. $folderN`
  * This lists the messages inside a folder.
* `maildir-tools message $file $file2 .. $fileN`
  * This formats and displays a single message.

Most of the sub-commands default to looking in `~/Maildir` but the `-prefix /path/to/root` will let you change the directory.  Maildirs are handled recursively, and things are pretty fast but I guess local SSDs help with that.  For everything else there is always the option to cache things.

(ProTip: The format-string you use makes a difference, for example `#{name}` is faster than `#{unread}`, which requires counting the messages which are unread.)


## Scripting Usage: Maildir List

The most basic usage is the following, which will recursively show the Maildirs beneath `~/Maildirs`:

`$ maildir-tools maildirs`

To change the output you can supply a format-string to specify what should be displayed, for example:

```
$ maildir-tools maildirs --format '#{name}' | grep debian-packages
..
/home/skx/Maildir/debian-packages
/home/skx/Maildir/debian-packages-Pkg-javascript-devel
/home/skx/Maildir/debian-packages-aatv
..
```

The following format-strings are available:

|     Flag |                                    Meaning  |
| -------- | ------------------------------------------- |
|     name | The name of the folder.                     |
|shortname | The name of the folder, without the prefix. |
|    total | The total count of messages in the folder.  |
|   unread | The count of unread messages in the folder. |




## Scripting Usage: Message List

The most basic usage is the following, which will output a summary of all the messages in the Maildir folder located at `~/Maildirs/example`:

`$ maildir-tools messages example`

To change the output you can supply a format-string to specify what should be displayed, for example:

```
$ maildir-tools messages --format '[#{4flags}] #{from} #{subject}' debian-packages-ttylog
..
[    S] Steve Kemp <skx@debian.org> Bug#553945: Acknowledgement (ttylog: Doesn't test length of device.)
..
```

The following format-strings are available:

|     Flag |                                   Meaning  |
| -------- | ------------------------------------------ |
|    flags | The flags of the message.                  |
|    file  | The filename of the message                |
|    index | The index of the message in the folder.    |
|    total | The total count of messages in the folder. |
| "header" | The content of the named header.           |

Headers are read flexibly, so if you used `#{subject}` the subject-header
would be returned.  Similar things work for all other headers.

There is one special case which is the support of displaying email
addresses specially.  For example if you had a message from: `"Steve Kemp" <steve@steve.fi>` then the contents of `#{from}` would be that string.

You might prefer to use:

* "`#{from.name}`" to receive the output "Steve Kemp".
* "`#{from.email}`" to receive the output "<steve@steve.fi>".



## Scripting Usage: Message Display

In the UI there is support for displaying a single message, via the use of:

`$ maildir-tools message path/to/maildir/cur/foo:2,S`

The output of this is currently hardwired, but in the future we might allow it to be changed via a template-file.


# Console Mail Client

To assume myself that the primitives are useful, and to have some fun I put together a simple console-based mail-client which you can invoke via:

```
maildir-tools ui
```

This is an __extremely__ minimal UI, which allows you to navigate:

* When launched you'll be presented with a list of Maildir folders.
* Scrolling up/down works as expected.
* Pressing return will open the maildir, replacing the view with a list of messages.
* Once again scrolling should be responsive and reliable.
* Finally you can view a message by hitting return within the list.

In each case you can return to the previous mode/view via `q`, or quit globally via `Q`.  When you're viewing a single message "`J`" and "`K`" move backwards/forwards by one message.

`vi` keys work, as do HOME, END, PAGE UP|DOWN, etc.
