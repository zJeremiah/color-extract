# Color Extractor

This simple app will print html text to stdout that will read an image and display it's pixels and the colors found in the image.

This is as simple as running:
```shell
go build
./color-extract test.png > test.html
```
* open the `test.html` file in a browser and it should show a breakdown of the image colors
* sorted by number of pixels from greatest to least
* Currently supports jpg, webp, and png

## Warning
If you use a large image file with this program, it will take a long time for the html file to load.

