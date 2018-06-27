# Gnuplot script: http://www.gnuplot.info/

set terminal png truecolor size 3000,2000 font 'Droid Sans Mono' 20 enhanced
set y2range[0:200]

set multiplot layout 3, 1
  set title "Writes vs Replicas"
  plot "scaler.dat" using 1:4 with lines title "writes" linecolor "#000000" axes x1y2, \
                "" using 1:2 with lines title "replicas" linecolor "#870202"

  set title "Writes vs Queue Length"
  plot "scaler.dat" using 1:4 with lines title "writes" linecolor "#000000" axes x1y2, \
                "" using 1:3 with lines title "queue length" linecolor "#bdd4f9"

  set title "Queue Length vs Replicas"
  plot "scaler.dat" using 1:3 with lines title "queue length" linecolor "#bdd4f9", \
                "" using 1:2 with lines title "replicas" linecolor "#870202"
unset multiplot