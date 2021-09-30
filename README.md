![](./static/header.png)

# mdcat

Go language remake of mdcat.

Uses the GitHub API to convert your markdown files to [GitHub styled](https://primer.style/) HTML site.

## Features

- Light/dark mode
- Code highlighting
- Web and mobile compatible display
- Meta information

## Install

- brew

```shell
brew install chyroc/tap/mdcat
```

- go

```shell
go install github.com/chyroc/mdcat@latest
```

## Usage

Usage is very simple:

```sh
mdcat <markdown_file.md>
```

It automatically generates HTML file in the same directory.


Default HTML Title is filename, you can add `--title` args to modify:

```shell
mdcat --title "Hi, Cat" <markdown_file.md>
```

Default output HTML file is `<input_filename.html>`, you can add `--output` args to modify:

```shell
mdcat --title "Hi, Cat" --output ./docs/index.html <markdown_file.md>
```

If the markdown file references another markdown file (in the form of [title](./markdown-file.md)), and you want to render the referenced file at the same time, then you can use the `--link` parameter:

```shell
mdcat --title "Hi, Cat" --output ./docs/index.html --link <markdown_file.md>
```

You can also use meta syntax to define the behavior of mdcat:

```markdown
--
title: "Hi Cat"
slug: index.html
--

Hello, World.
```

This Meta is like command: `mdcat --title "Hi, Cat" --output ./index.html <markdown_file.md>`.

## Demo

You can see this markdown file's HTML on:
[here](https://chyroc.github.io/mdcat)

## Thanks

- Thanks for py version: https://github.com/calganaygun/MDcat.
- Thanks to [Karma](https://www.instagram.com/sanmiyorumamaevet/) for the cat illustration in the header.
