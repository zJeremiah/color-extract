#!/usr/local/bin/bash

# for resizing an image and indexing using only 20 colors
magick input.png -resize 500x -colors 20 output.png 

# don't resize, just index limited number of colors
magick input.png -colors 20 output.png

# for converting the grid image output to a density that can be printed with detail (and the correct print size)
magick -units PixelsPerCentimeter -density 150 input.grid.png output.grid.png