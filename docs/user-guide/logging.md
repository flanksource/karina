`co.elastic.logs/enabled`
Filebeat gets logs from all containers by default, you can set this hint to false to ignore the output of the container. Filebeat won’t read or send logs from it. If default config is disabled, you can use this annotation to enable log retrieval only for containers with this set to true.

```yaml
co.elastic.logs/multiline.*
co.elastic.logs/multiline.pattern: '^\['
co.elastic.logs/multiline.negate: true
co.elastic.logs/multiline.match: after
```

Multiline settings. See Multiline messages for a full list of all supported options.

`co.elastic.logs/json.*`
JSON settings. See json for a full list of all supported options.

`co.elastic.logs/include_lines`
A list of regular expressions to match the lines that you want Filebeat to include. See Inputs for more info.

`co.elastic.logs/exclude_lines`
A list of regular expressions to match the lines that you want Filebeat to exclude. See Inputs for more info.

`co.elastic.logs/module`
Instead of using raw docker input, specifies the module to use to parse logs from the container. See Modules for the list of supported modules.

`co.elastic.logs/fileset`
When module is configured, map container logs to module filesets. You can either configure a single fileset like this:

`co.elastic.logs/processors`
Define a processor to be added to the Filebeat input/module configuration. See Processors for the list of supported processors.

In order to provide ordering of the processor definition, numbers can be provided. If not, the hints builder will do arbitrary ordering:

```yaml
co.elastic.logs/processors.1.dissect.tokenizer: "%{key1} %{key2}"
co.elastic.logs/processors.dissect.tokenizer: "%{key2} %{key1}"
````
In the above sample the processor definition tagged with 1 would be executed first.
