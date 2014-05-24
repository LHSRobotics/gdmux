G21 ; set units to millimeters
G1 X0 Y-25 Z20


G1 Z10

G2  Y25          I0  J25  K0
G2  Y-25         J-25

G1  Z20

G1  Y-13
G1  Z10

G2  Y13          I0  J13  K0

G1 Z20

G1  X16  Y8

G1 Z10

G2  Y12          J2
G2  Y8           J-2

G1 Z20



G1  X16  Y-8

G1 Z10

G3  Y-12          J-2
G3  Y-8           J2


G1 Z20
