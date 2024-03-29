#!/usr/bin/env python3

import json
import re
import sys

api = json.load(open("cmc.json"))


def typetr(t):
    if t == "integer":
        return "int64"
    if t == "boolean":
        return "bool"
    if t == "number":
        return "float64"
    return t


def unfold(d):
    """
    Unfold a dictionary with keys representing paths into a nested dictionary structure.

    :param d: A dictionary with keys as paths (like 'a/b/c') and values as the end values.
    :return: A nested dictionary representing the unfolded structure.
    """
    unfolded = {}

    for key, value in d.items():
        # Split the key into parts
        parts = key.split('/')

        # Reference to the current level in the unfolded dictionary
        current_level = unfolded

        for part in parts[:-1]:  # iterate through the parts except the last one
            # If the part is not yet a key in the current level, add it as an empty dict
            if part not in current_level:
                current_level[part] = {}
            # Move deeper into the unfolded structure
            current_level = current_level[part]

        # Assign the value to the last part
        current_level[parts[-1]] = value

    return unfolded


paths = []
dpaths = {}

print("package main\n\n")
print("// This file is generated by gen.py. DO NOT EDIT.\n\n")
print("var apiMap = map[string]string{")
for k, v in api.items():
    typename = "".join([i.replace("-", "").title()
                       for i in re.split("[-/]", k)[1:]])
    short = "/".join(k.split("/")[1:])
    print('    "{}": "{}",'.format(short, typename))
    paths.append(short)
    pdict = dict(url=k, opid=typename, params=v)
    dpaths[short] = pdict
print("}\n\n")


def emit(key, dic, pre):
    if pre == "":
        print("var api = &ApiNode{")
    else:
        print(pre + "{")
    print(pre + '    Label: "{}",'.format(key))

    if "opid" in dic:
        ps = dic["params"]
        print(pre + '    Url: "{}",'.format(dic["url"]))
        print(pre + '    Args:  map[string]Arg{')
        for p in ps:
            if 'defa' not in p:
                p['defa'] = ""
            print(pre + '         "{}": {{"{}", "{}", "{}", "{}"}},' .format(
                p['name'], p['name'], p["type"], p['desc'].replace('"', '\\"'),
                p['defa'].replace('"', '\\"')))
        print(pre + "    },")
    else:
        print(pre + '    Children: []*ApiNode{')
        for k, v in dic.items():
            emit(k, v, pre + "        ")
        print(pre + "    },")
    if pre == "":
        print("}")
    else:
        print(pre + "},")


uf = unfold(dpaths)
emit("root", uf, "")

sys.exit(0)

for k, v in api.items():
    typename = "".join([i.replace("-", "").title()
                       for i in re.split("[-/]", k)[1:]])
    print("type ", typename, "struct {")
    for i in v:
        orign = i['name']
        nam = "".join([i.title() for i in i['name'].split("_")])
        typ = typetr(i['type'])
        des = i['desc']
        tag = 'arg:"" name:"{}" optional:"" help:"{}'.format(
            orign, des.replace('"', '\\"'))
        if 'defa' in i:
            tag += ' Default: {}"'.format(i['defa'].replace('"', '\\"'))
        else:
            tag += '"'
        l = "    {} {} `{}`".format(nam, typ, tag)
    print(l)
    print('    Path string `arg:"" optional:"" name:"path" default:"{}"`'.format(k))
    print("}\n")
