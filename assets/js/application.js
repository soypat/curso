require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
// require("@fortawesome/fontawesome-free/js/all.js");
// require("popper.js/dist/popper.min.js");
// require("bootstrap/dist/js/bootstrap.min.js");


// function executes upon loading all of DOM

$(() => {
    // window.onload = function() {
    let hldivs = document.querySelectorAll("div.highlight")
    hldivs.forEach(function (div) {
        let hyphIdx = div.className.indexOf("highlight-");
        let langName = div.className.substring(hyphIdx+"highlight-".length);
        let endLangIdx = langName.indexOf(" ")
        if (endLangIdx>0) {
            langName = langName.substring(0,endLangIdx)
        }
        let pre = div.firstChild
        pre.setAttribute("class","hljs "+langName)
        div.setAttribute("class","")//"hljs highlight "+langName) // erase autodetected lang
    })
    document.querySelectorAll('code').forEach((block) => {
        let attr = block.getAttribute("class")
        if (!block.hasAttribute("class") || attr === "highlight" || attr === "col-md-7") {
            block.setAttribute("class","language-python");
        }
    });
    render(RENDERERS);
});

RENDERERS.push(renderCodeBlock)
function renderCodeBlock() {
    document.querySelectorAll('code , .highlight, .hljs').forEach((block) => {
        hljs.highlightBlock(block);
    });

}

RENDERERS.push(readyJSCodeBlock)
// This function prepares javascript generated
// markdown for highlight.js rendering. front-end
// devs don't agree on anything, apparently
function readyJSCodeBlock() {
    let preCods = document.querySelectorAll("pre > code")
    preCods.forEach(function (cod) {
        let hyphIdx = cod.className.indexOf("language-");
        let langName = cod.className.substring(hyphIdx + "language-".length);
        let endLangIdx = langName.indexOf(" ")
        if (endLangIdx > 0) { // discard classes after language name
            langName = langName.substring(0, endLangIdx)
        }
        cod.setAttribute("class", "hljs " + langName)
    })
}

// Use tabs in content/markdowny textareas
$(document).delegate('#content', 'keydown', function (e) {
    var keyCode = e.keyCode || e.which;
    if (keyCode == 9) {
        e.preventDefault();
        var start = this.selectionStart;
        var end = this.selectionEnd;
        // set textarea value to: text before caret + tab + text after caret
        $(this).val($(this).val().substring(0, start)
            + "\t"
            + $(this).val().substring(end));
        // put caret at right position again
        this.selectionStart =
            this.selectionEnd = start + 1;
    }
});