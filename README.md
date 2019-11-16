# wfsports: WeFunder Sports challenge

Start by running
```
$ wfsports start names.csv
```
This outputs a file called round1.csv.
Each row of this file has the names of two players competing in this round.
As the organizer, you get to add the name of the winner to each of the rows
once the round is done. If the number of players isn't a power of two,
there will be some players who get a free pass to go to the next round.
They will show up in rows like `Alice,Alice,Alice`, meaning that Alice
has played against her self and won.

After entering the winners in round1.csv, run
```
$ wfsports next round1.csv
```
That will make random pairs of winners from round 1 and output round2.csv
in the same format as round1.csv.
Once they play each other, you as the organizer get to add the winners'
names as in round1, and repeat running `wfsports next round2.csv` and
repeat the process until someone wins the whole tournament.

To display the players for round N on the projector, run this (on Mac):
```
$ wfsports show roundN.csv
$ open table.html
```

