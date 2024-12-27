<h1 align="center">
  crossplane-trace-explorer
</h1>

<p align="center">
  🧰 Enhanced Crossplane trace explorer
</p>

![screenshot](./screenshot.png)

`crossplane trace` is a very handy tool, but it is not very interactive and requires a few extra
hops to properly debug its traced objects. This tool aims on closing this gap by providing
an interactive tracing explorer based on the tool tracer output.

## ✨ Features

- Allow copying/yanking object value, including API group to avoid conflict
- Display last transition times by default

### Upcoming

- Describe Kubernetes object from the explorer
- Allow expanding error messages in the UI (shortcut `K`)
- Allow mutating resource annotations (pause, finaliser)

## 📀 Install

### Linux and Windows

[Check the releases section](https://github.com/brunoluiz/crossplane-trace-explorer/releases) for more information details.

### MacOS

```
brew install brunoluiz/tap/crossplane-trace-explorer
```

### Other

Use `go install` to install it

```
go install github.com/brunoluiz/crossplane-trace-explorer@latest
```

## ⚙️ Usage

You must have `crossplane` installed. Run the tracer with `-o json` and pipe it to this tool.

```
crossplane beta trace Bucket/test-resource-bucket-hash -o json | crossplane-trace-explorer
```

## 🧾 To-do

- Understand how to add a viewport within the explorer to expand leafs
- Re-do the `addNodes` feature
- Use a map instead of slice on the `tree#Node.details` (have to be careful around ordering)
- Add comments to functions
- Review hooks x Elm Bubble model
- Allow short and wide table format
- Understand why first render of statusbar is not rendering selected path without hack
- Add support to logging through TUI debug mode to file
