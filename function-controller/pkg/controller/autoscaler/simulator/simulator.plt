# Gnuplot script: http://www.gnuplot.info/

set terminal png truecolor size 2000,2000 font 'Droid Sans Mono' 20 enhanced

plot "scaler.dat" using 1:2 with lines title "instances", "" using 1:3 with lines title "queue length"
