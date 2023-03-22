### `delete()` {#fn-delete}
[:octicons-tag-24: Version-1.5.8](../changelog.md#cl-1.5.8)

函数原型：`fn delete(src: map[string]any, key: str)`

函数说明： 删除 json map 中的 key

```python

# input
# {"a": "b", "b":[0, {"c": "d"}], "e": 1}

# script
j_map = load_json(_)

delete(j_map["b"][-1], "c")

delete(j_map, "a")

add_key("j_map", j_map)

# result:
# {
#   "j_map": "{\"b\":[0,{}],\"e\":1}",
# }
```