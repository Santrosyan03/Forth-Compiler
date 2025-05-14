variable a
5 a !
1 dup .s
a @ .
10 3 mod .
1 2 tuck .s
variable x
x @ .
3 b !
b @ .
11 2 over .s
variable b
42 x !
7 3 * .
b @ .
0 5 mod .
x @ .
5 3 + .
1 2 nip .s
a @ b @ + .
1 2 drop .s
10 5 - .
b @ a @ - .
1 2 swap .s
5 neg .
.s
