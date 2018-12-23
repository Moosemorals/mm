/* 
 * Copyright 2015 Osric Wilkinson <webmoose@fluffypeople.com>
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
 * SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION
 * OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN
 * CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 *
 */

var constraints = {};
var words = [];
var avgWordLen = 0;
var xmlExtra = "<>&\"'";
var numberSymbols = "0123456789";
var symbols = "!#$%()*+,-./-:;=?@[\\]^_`{|}~";

var crypto = window.crypto || window.mscrypto || null;

function addClass(targetId, className) {
    document.getElementById(targetId).classList.add(className);
}

function removeClass(targetId, className) {
    document.getElementById(targetId).classList.remove(className);
}

function hasClass(targetId, className) {
    return document.getElementById(targetId).classList.contains(className);
}

function setText(targetId, text) {
    var node = document.getElementById(targetId);

    while (node.hasChildNodes()) {
        node.removeChild(node.firstChild);
    }
    node.appendChild(document.createTextNode(text));
}


function getRandomNumber() {
    if (crypto !== null) {
        var r = new Uint32Array(1);
        crypto.getRandomValues(r);
        return r[0] / Math.pow(2, 32);
    } else {
        return Math.random();
    }
}

function randomInt(max) {
    return Math.floor(getRandomNumber() * max);
}

function titleCase(word) {
    return word.substring(0, 1).toUpperCase() + word.substring(1);
}

function loadWords(filename) {
    var client = new XMLHttpRequest();
    client.open('GET', filename);
    client.onreadystatechange = function () {
        if (client.readyState === 4) {
            words = client.responseText.split("\n");

            var totalWordLength = 0;
            for (var i = 0; i < words.length; i += 1) {
                totalWordLength += words[i].length;
            }
            avgWordLen = totalWordLength / words.length;
            generate();
        }
    };
    client.send();
}

function getNumber(targetId) {
    var raw = document.getElementById(targetId).value;

    if (typeof raw !== 'undefined') {
        // has some kind of value
        return parseInt(raw, 10);
    } else {
        return Math.NaN;
    }
}

function setNumber(targetId, number) {
    document.getElementById(targetId).value = isNaN(number) ? "" : number;
}

function getValidWord() {
    var i, letters, word;

    if (!constraints['mayLetters'] && !constraints['mayCapitals']) {
        // If the user doesn't want upper or lower case letters, constuct
        // a random 'word' from numbers and/or symbols.
        var symbolSet = "";
        if (constraints['mayNumbers']) {
            symbolSet += numberSymbols;
        }
        if (constraints['maySymbols']) {
            symbolSet += symbols;
            if (constraints['xmlSafe']) {
                symbolSet += xmlExtra;
            }
        }
        word = "";
        for (i = 0; i < avgWordLen; i += 1) {
            word += symbolSet[randomInt(symbolSet.length)];
        }
    } else {
        // Pull a word from the list. Make sure that if they don't want
        // numbers or symbols, it doesn't contain them.
        // TODO: Check that there are valid 'words' in the word list, or
        // this loops forever. (Statisticaly, it will loop forever occasioanlly)
        do {
            word = words[randomInt(words.length)];
            if (word === "") {
                continue;
            }
            if (!constraints['mayNumbers'] && word.match(/[0-9]/)) {
                continue;
            }
            if (!constraints['maySymbols'] && word.match(/[^a-zA-Z0-9]/)) {
                continue;
            }
            break;
        } while (true);
    }

    if (constraints['mustCapitals']) {
        // If the user wants capitals, give them capitals...
        letters = word.split("");
        word = "";
        for (i = 0; i < letters.length; i += 1) {
            if (randomInt(2) === 1) {
                word += letters[i].toUpperCase();
            } else {
                word += letters[i].toLowerCase();
            }
        }
    } else if (constraints['mayCapitals'] && randomInt(2) === 1) {
        word = titleCase(word);
    }

    if (constraints['mustNumbers'] && !word.match(/[0-9]/)) {
        word += numberSymbols[randomInt(numberSymbols.length)];
    }

    if (constraints['mustSymbols'] && !word.match(/[^a-zA-Z0-9]/)) {
        word += symbols[randomInt(symbols.length)];
    }

    if (!constraints['mayCapitals'] && constraints['mayLetters']) {
        word = word.toLowerCase();
    } else if (!constraints['mayLetters'] && constraints['mayCapitals']) {
        word = word.toUpperCase();
    }

    return word;
}

function getSeparator() {
    var separators = document.getElementById("controls").elements['separator'];

    for (var i = 0; i < separators.length; i += 1) {
        if (separators[i].checked) {
            return separators[i].value;
        }
    }
}

function generate() {
    if (words.length === 0) {
        setText("password", "Wordlist still loading");
        return;
    }

    if (!constraints['mayLetters'] && !constraints['mayCapitals'] && !constraints['mayNumbers'] && !constraints['maySymbols']) {
        setText("password", "Nice try, but you must select something to get a password.");
        return;
    }

    var length = getNumber("count");
    if (!isNaN(length)) {
        // Length is valid. Make sure its more than zero.
        if (length < 1) {
            length = 1;
            setNumber("count", length);
        }
    } else {
        // Nothing is valid. default length of 6
        length = 6;
    }

    var word, i, result = "",
        symbolSet;

    if (constraints['maySymbols']) {
        symbolSet += symbols;
        if (constraints['xmlSafe']) {
            symbolSet += xmlExtra;
        }
    }

    for (i = 0; i < length; i += 1) {
        word = getValidWord();

        if (i < (length - 1)) {
            switch (getSeparator()) {
                case "space":
                    word += " ";
                    break;
                case "symbol":
                    word += symbolSet[randomInt(symbolSet.length)];
                    break;
                case "none":
                    //dropthrough
                default:
                    // do nothing
                    break;
            }

        }

        result += word;
    }

    var maxLen = getNumber("maxLen");
    if (!isNaN(maxLen)) {
        result = result.substring(0, maxLen);
    }

    setText("password", result);
}

function toggleVisible(target) {
    if (hasClass(target, "hidden")) {
        removeClass(target, "hidden");
    } else {
        addClass(target, "hidden");
    }
}

// Run at load time
function init() {
    // Re-generate password when form elements change
    ["count", "xmlSafe", "maxLen"].forEach(function (entry) {
        document.getElementById(entry).addEventListener("change", generate, false);
    });

    var xmlSafe = document.getElementById("xmlSafe");
    constraints['xmlSafe'] = xmlSafe.checked;
    xmlSafe.addEventListener("click", function () {
        constraints['xmlSafe'] = xmlSafe.checked;
    });

    var sepSymbol = document.getElementById("sepSymbol");
    document.getElementById("maySymbols").addEventListener("click", function () {
        xmlSafe.disabled = !this.checked;
        sepSymbol.disabled = !this.checked;
        if (!this.checked) {
            xmlSafe.checked = false;
            sepSymbol.checked = false;
            document.getElementById("sepSpace").checked = true;
        }
    }, false);

    ["Space", "Symbol", "None"].forEach(function (entry) {
        document.getElementById("sep" + entry).addEventListener("click", generate, false);
    });

    // Fill constraints array
    // Untick and diable 'must' when 'may' is unticked
    // Re-enable 'must' when 'may' is ticked
    ["Letters", "Capitals", "Numbers", "Symbols"].forEach(function (entry) {
        var may = document.getElementById("may" + entry);
        var must = document.getElementById("must" + entry);

        constraints["may" + entry] = may.checked;
        constraints["must" + entry] = must.checked;

        may.addEventListener("click", function () {
            must.disabled = !may.checked;
            if (!may.checked) {
                must.checked = false;
                constraints["may" + entry] = false;
                constraints["must" + entry] = false;
            } else {
                constraints["may" + entry] = true;
            }
            generate();
        }, false);
        must.addEventListener("click", function () {
            constraints["must" + entry] = must.checked;
            generate();
        }, false);
    });


    // Do calculation on button click
    document.getElementById("generate").addEventListener("click", generate, false);


    if (crypto === null) {
        toggleVisible("warning");
    }

    loadWords("/js/dicelist.txt");
    generate();
}

window.addEventListener("load", init, false);
