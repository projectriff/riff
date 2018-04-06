# Gnuplot script: http://www.gnuplot.info/

set terminal png truecolor size 2000,2000 font 'Droid Sans Mono' 20 enhanced
# set logscale y 2

plot "scaler.dat" using 1:4 with lines title "writes" linecolor "#000000", "" using 1:3 with lines title "queue length" linecolor "#bdd4f9", "" using 1:2 with lines title "replicas" linecolor "#870202"
