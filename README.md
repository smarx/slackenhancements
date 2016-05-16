# Slack Enhancements
A "[Stupid hackathon](https://stupidhackathon.github.io)" project. It solves the problem of your important messages being missed by your coworkers. See https://slackenhancements.site44.com for a demo presentation. (Note that the `<important>` tag no longer exists!)

```
 _________________________
< SLACK ENHANCEMENTS      >
 -------------------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```

## Setup
1. If you don't have go yet, install it.
2. `go get github.com/smarx/go-cowsay github.com/nlopes/slack`

## Running
`go run enhance.go`

If you haven't yet, that will tell you how to get an API token for your account and pass it to the app.

## Usage
Available tags are `<blink>`, `<marquee>`, and `<cow>`.

Balanced tags are required and apply to the entire message, regardless of placement. The following are all equivalent:
* `<blink><marquee>Hello!</marquee></blink>`
* `<blink><marquee>Hello!</blink></marquee>`
* `<blink>Hel</blink>lo!<marquee></marquee>`
