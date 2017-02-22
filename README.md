# kolorit

[![Circle CI](https://circleci.com/gh/ktat/kolorit/tree/master.svg?style=shield)](https://circleci.com/gh/ktat/kolorit/tree/master)

coloring text with regexp

![](https://raw.githubusercontent.com/ktat/kolorit/master/kolorit.gif)

# USAGE

```
usage: kolorit [-f file|-[rgbycpwk] regexp|-f pattern|-R dir|-h]  [file ..]

        -f file name pattern ... file pattern. read from matched file
        -R dir  ... recursively read directory
        -r   regexp ... to be red
        -g   regexp ... to be green
        -b   regexp ... to be blue
        -y   regexp ... to be yellow
        -c   regexp ... to be cyan
        -p   regexp ... to be purple
        -w   regexp ... to be white
        -k   regexp ... to be black
        -lb  regexp ... to be light_blue
        -lc  regexp ... to be light_cyan
        -lg  regexp ... to be light_green
        -lgr regexp ... to be light gray
        -lp  regexp ... to be light_purple
        -lr  regexp ... to be light_red
        -ly  regexp ... to be light_yellow
        -e   regexp ... erace matched string
        -br   color_name ... background color of red
        -bg   color_name ... background color of green
        -bb   color_name ... background color of blue
        -by   color_name ... background color of yellow
        -bc   color_name ... background color of cyan
        -bp   color_name ... background color of purple
        -bw   color_name ... background color of white
        -bk   color_name ... background color of black
        -blb  color_name ... background color of light_blue
        -blc  color_name ... background color of light_cyan
        -blg  color_name ... background color of light_green
        -blgr color_name ... background color of light gray
        -blp  color_name ... background color of light_purple
        -blr  color_name ... background color of light_red
        -bly  color_name ... background color of light_yellow
        -s ... regexp option. tread given content as single line(default as multi line)
        -I ... set matched text background color inverted color
        -nI ... ignore -I option
        -B ... matched text to be bold
        -nB ... ignore -B option
        -m ... regexp for multiline
        -i ... regexp is case insensitive
        -d ... print debug message
        -h ... help
        -help ... help
        -use ... use predefined setting from config file($HOME/.koloit.toml)
        -conf ... path of config file (default "$HOME/.kolorit.toml")
        -grep ... take string and ignore not matched lines with it like grep. cannot use it with -s option
        -ngrep ... ignore grep option.
        -and  ... change grep option behavior. take string only when all regexps are match.
```
# Config file

You can predefine color regexp, B, s, m, i, e, grep and ng options in config file($HOME/.kolorit.toml) like the following
```
[default]
# specify default kolorit behabior
B = true

[calc]
y = '[=?.<>\-+*/]+'
b = '\d+'

[date_time]
y = '\d{4}[/-]\d{2}[-/]\d{2}'
b = '\d{2}:\d{2}(?::\d{2})?'

[rsync]
g = 'sending incremental file list(.+?)\nsent [\d.]+\w bytes' 
s = true
B = true
```

and you can use it like:
```
% kolorit -use calc -f one.txt
% echo "2017-01-01 10:00:00" | kolorit -use date_time
```
# Example

```
% rsync -avhn /tmp/a/ /tmp/b/ | kolorit -r '\w+' -p '\d+'
% rsync -avhn /tmp/a/ /tmp/b/ | kolorit -s -g 'sending incremental file list(.+?)\nsent [\d.]+\w bytes'
% rsync -avhn /tmp/a/ /tmp/b/ | kolorit -use rsync
% godoc time |kolorit -r 'current|local' -y 'reference time' | less -R
% godoc time |kolorit -B -r 'current|local' -y 'reference time' --grep 
```

# Author

Atsushi Kato (ktat)

# License

MIT: https://ktat.mit-license.org/2016