# kolorit

[![Circle CI](https://circleci.com/gh/ktat/kolorit/tree/master.svg?style=shield)](https://circleci.com/gh/ktat/kolorit/tree/master)

coloring text with regexp

![](https://raw.githubusercontent.com/ktat/kolorit/master/kolorit.gif)

# Usage

```
  kolorit [options] [FILES]
  kolorit [options] -f "*.go"
  kolorit [options] -R [FILES/DIRECTORIES]
```

# Options
```
  -help
        show usage
  -h    show usage
  -conf string
        path of config file
  -use string
        use predefined setting from config file($HOME/.kolorit.toml)
  -grep
        take string and ignore not matched lines with it like grep. cannot use it with -s
  -and
        change grep option behavior. take string only when all regexps are matched.
  -ngrep
        ignore grep option
  -s    regexp option. treat given content as single line(default as multi line)
  -i    regexp option. do case insensitive pattern matching.
  -R    recursively read directory.
  -f string
        file pattern. read from matched file.
  -e string
        erase matched string
  -B    matched string to be bold
  -nB
        ignore -B option
  -I    matched string background color to be inverted
  -nI
        ignore -I option
  -dot
        dot includes files starts with '.'
  -vcs
        vcs includes vcs files/dirs
  -d    debug mode
```
# Color Options:
```
  -r regexp     to be red
  -g regexp     to be green
  -b regexp     to be blue
  -y regexp     to be yellow
  -p regexp     to be purple
  -c regexp     to be cyan
  -k regexp     to be black
  -w regexp     to be white
  -lr regexp    to be light_red
  -lg regexp    to be light_green
  -lb regexp    to be light_blue
  -ly regexp    to be light_yellow
  -lp regexp    to be light_purple
  -lc regexp    to be light_cyan
  -dgr regexp   to be dark_gray
  -lgr regexp   to be light gray
```
# Back Ground Color Options:
* color_name is name of color explained the above
```
  -br color_name        background color of red
  -bg color_name        background color of green
  -bb color_name        background color of blue
  -by color_name        background color of yellow
  -bp color_name        background color of purple
  -bc color_name        background color of cyan
  -bk color_name        background color of black
  -bw color_name        background color of white
  -blr color_name       background color of light_red
  -blg color_name       background color of light_green
  -blb color_name       background color of light_blue
  -bly color_name       background color of light_yellow
  -blp color_name       background color of light_purple
  -blc color_name       background color of light_cyan
  -bdgr color_name      background color of dark_gray
  -blgr color_name      background color of light gray
```
# Config file

You can predefine color regexp, B, s, m, i, e, grep and ngrep options in config file($HOME/.kolorit.toml) like the following
```
[default]
# specify default kolorit options
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