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
        --grep ... take string and ignore not matched lines with it like grep
```

# License

MIT

# Author

Atsushi Kato (ktat)

