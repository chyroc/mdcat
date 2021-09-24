![](./static/header.png)

# mdcat

Go language remake of mdcat.

Uses the GitHub API to convert your markdown files to [GitHub styled](https://primer.style/) HTML site.

## Features

- Light/dark mode
- Code highlighting

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

Default HTML Title is filename, you can add `--title` args to modify:

```shell
mdcat --title "Hi, Cat" <markdown_file.md>
```

It automatically generates HTML file in the same directory.

## Demo

You can see this markdown file's HTML on:
[here](https://chyroc.github.io/mdcat/README.html)

## Thanks

- Thanks for py version: https://github.com/calganaygun/MDcat.
- Thanks to [Karma](https://www.instagram.com/sanmiyorumamaevet/) for the cat illustration in the header.
