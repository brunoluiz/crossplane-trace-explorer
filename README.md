<h1 align="center">
  crossplane-trace-explorer
</h1>

<p align="center">
  🧰 Enhanced Crossplane trace explorer
</p>

`crossplane trace` is a very handy tool, but it is not very interactive and requires a few extra
hops to properly debug its traced objects. This tool aims on closing this gap by providing
an interactive tracing explorer based on the tool tracer output.

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

## Future features

- Describe Kubernetes object from the explorer
- Allow expanding error messages in the UI (shortcut `K`)
- Allow mutating resource annotations (pause, finaliser)
- Reorganise code around tree to allow more modals
