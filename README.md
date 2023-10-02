# Unofficial crossplane CLI - cp-cli
This is a hobby project I built as I wanted to gain more experience in Go and especially the k8s clients. The goal was to implement two commands `describe` and `diagnose`.

# Commands
## describe
The describe command takes a Composite Resource or Claim resource and name of the resource as args input. It then gets the resource and all its children and prints it out either as table in the CLI or a .png.


| Variable Name  | Shorthand | Default   | Description                                                                                           |
|----------------|-----------|-----------|-------------------------------------------------------------------------------------------------------|
| namespace      | -n        | "default" | Kubernetes namespace                                                                                  |
| kubeconfig     | -k        | ""        | Path to the Kubeconfig file.                                                                         |
| output         | -o        | "cli"     | Output format of the resource. Must be one of "cli" or "graph".                                      |
| fields         | -f        | parent, kind, name, synced, ready   | Comma-separated list of fields to display. Available fields are "parent", "name", "kind", "namespace", "apiversion", "synced", "ready", "message", "event". |
| path           | -p        | "./graph.png" | Absolute path and filename for the output graph PNG. The filename must end with '.png'.             |

**Usage:** cp-cli describe TYPE[.GROUP] NAME 

**Example usage:**
1. cp-cli describe objectstorage my-object-storage
or
2. cp-cli describe objectstorage my-object-storage -f name,kind,apiversion -o graph 

## diagnose
The diagnose command takes a Composite Resource or Claim resource and name of the resource as args input. Then some health checks are performed on the resource and its children. Every resource that is considered unhealthy will be printed out. 

| Variable Name  | Shorthand | Default   | Description                                                                                           |
|----------------|-----------|-----------|-------------------------------------------------------------------------------------------------------|
| namespace      | -n        | "default" | Kubernetes namespace                                                                                  |
| kubeconfig     | -k        | ""        | Path to the Kubeconfig file.                                                                         |

**Usage:** cp-cli describe TYPE[.GROUP] NAME 

**Example usage:**
1. cp-cli diagnose objectstorage my-object-storage
or
2. cp-cli diagnose objectstorage my-object-storage -n my-namespace



# Reference
cp-cli has been inspired by other projects:
- https://github.com/crossplane/crossplane-cli (Old and archived crossplane-cli by crossplane)
- https://github.com/tohjustin/kube-lineage 