#gdmux

Invocation: `gdmux -arm /dev/ttyStaubli -extruder /dev/ttyExtruder`

Since this will mainly be running on Linux, we just deal with the serial ports as files.
It's up to the user to set them up with the correct parameters (baudrate, stop bits, parity, etc.) using `stty`.
This keeps things nice and simple.
