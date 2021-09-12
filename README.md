# GuitarScales
_A program that finds "optimal" ways to play minor/major scales on the guitar fretboard._
# Usage
```shell
go build generate_png_images_for_scales.go
./generate_png_images_for_scales -octave=<OCTAVE>
```
where _<OCTAVE>_ is an integer between 2 and 6 (both incl.).

To generate everything use:
```shell
go build generate_png_images_for_scales.go
seq 2 6 | xargs -I {} ./generate_png_images_for_scales -octave={}
```

# Examples
![C3maj](png/C3maj/0.png?raw=true "C3maj")
![D#3min](png/D%233min/0.png?raw=true "C3maj")
![E3maj](png/E3maj/7.png?raw=true "C3maj")
