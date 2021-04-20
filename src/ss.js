let s = [

    [0, 1, 1, 0],
    [1, 1, 1, 1],
    [1, 1, 1, 1],
    [1, 1, 0, 0]

];



let res = [];

let potential = 0;
for (i = 0; i < s.length; i++) {
    res[i] = [];
    for (j = 0; j < s[i].length; j++) {
        if (s[i][j] == 1) {
            if (res[i].length == 0) {
                res[i].push(j);
                res[i][1] = 0;
            } else {
                res[i][1]++;
            }
        } else {
            if (res[i][0] != undefined) {
                break;
            }
        }

    }
}

let pot;
let little;
for (g = 0; g < res.length; g++) {
    if (res[g].length != 0) {
        pot = res[g];
        little = pot[1];
        for (h = 0; h < res.length; h++) {
            if (res[h].length != 0) {
                if (pot[0] == res[h][0] && res[h][1] <= pot[1]) {
                    if (res[h][1] < little && res[h][1] != res[h][0]) {
                        little = res[h][1];
                    }
                    debugger;

                    if (res[h][1] == res[h][0]) {
                        continue;
                    }

                    if (((h - g + 1) * (little - pot[0] + 1)) > potential) {
                        potential = ((h - g + 1) * (little - pot[0] + 1));
                    }
                }
            }
        }
    }
}

console.log(potential);