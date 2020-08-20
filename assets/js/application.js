require("expose-loader?$!expose-loader?jQuery!jquery");
require("bootstrap/dist/js/bootstrap.bundle.js");
require("@fortawesome/fontawesome-free/js/all.js");

$(() => {

});

// var outputResponse = new XMLHttpRequest();
// document.querySelector("#submit").addEventListener("click", function () {
//     outputResponse.open("POST", "<%= base %>", true);  // this is a plush template!
//     outputResponse.setRequestHeader("Content-Type", "application/json");
//     outputResponse.send(JSON.stringify({
//         code: codeID.value
//     }))
//     outputID.innerHTML = ""
// });