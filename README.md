# go-coloring

coloring text with regexp

# USAGE

```
usage: coloring [-f file|-[rgbycpwk] regexp|-f pattern|-R dir|-h]  [file ..]

        -f file_name/pattern/-(stdin) ... read from file. read stdin if '-' is given
        -R dir  ... recursively read directory
        -r regexp ... to be red
        -g regexp ... to be green
        -b regexp ... to be blue
        -y regexp ... to be yellow
        -c regexp ... to be cyan
        -p regexp ... to be purple
        -w regexp ... to be white
        -k regexp ... to be black
        -e regexp ... erace matched string
        -m ... regexp for multiline
        -i ... regexp is case insensitive
        -P ... use builtin pager
        -h ... help
        -d ... print debug message
        -use ... use predefined setting from config file($HOME/.koloit.toml)
        -conf ... path of config file (default "/home/ktat/.kolorit.toml")
        -grep ... take string and ignore not matched lines with it like grep
```
# Config file

You can predefine color regexp in config file($HOME/.kolorit.toml) like the following
```
[calc]
y = '[=?.<>\-+*/]+'
b = '\d+'

[date_time]
y = '\d{4}[/-]\d{2}[-/]\d{2}'
b = '\d{2}:\d{2}(?::\d{2})?'
```

and you can use it like:
```
% kolorit -use calc -f one.txt
% echo "2017-01-01 10:00:00" | kolorit -use date_time
```

# Author

Atsushi Kato (ktat)

# License

MIT: https://ktat.mit-license.org/2016