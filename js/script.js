$(document).ready(function() {
    $("#landing").css('height', ($(window).height()-58).toString());
    $("#survey-background").css('height', ($(window).height()-58).toString());
    $(window).resize(function() {
        $("#landing").css('height', ($(window).height()-58).toString());
        $("#survey-background").css('height', ($(window).height()-58).toString());
    });
    $("#register").on('click', function(e) {
        e.preventDefault();
    });
    $("#login-submit").on('click', function(e) {
        e.preventDefault();
        $.ajax({
            url: '/login',
            type: 'post',
            dataType: 'html',
            data: {
                username: $("#login-username").val(),
                password: $("#login-password").val()
            },
            success: function(data) {
                console.log(data);
                if (data === "true") {
                    window.location.href = "/";
                } else {
                    $("#login-error-message").show();
                }
            },
        });
    });
    $("#start-survey-button").on('click', function(e) {
        e.preventDefault();
        var user = getCookie("session-id");
        if (user === "") {
            $(".alert").show();
        } else {
            window.location.href = "/survey";
        }
    });
    $(".close").on('click', function(e) {
        $(".alert").hide();
    });
    $("#survey button").on('click', function(e) {
        var buttonId = event.target.id;
        var id = buttonId.split("-")[1];
        var currentQuestion = getCookie("current-question");
        var questionIndex = getCookie("current-qindex");
        var index = parseInt(questionIndex) + 1;
        var questionNumber = parseInt(currentQuestion) + 1;
        $.ajax({
            url: '/api/recordUserResponse',
            type: 'post',
            dataType: 'html',
            data: {
                response: id
            },
            success: function(data) {
                if (data === "true") {
                    currentIndex = index.toString();
                    currentQuestion = questionNumber.toString();
                    setCookie("current-qindex", currentIndex, 1);
                    setCookie("current-question", currentQuestion, 1);
                    if (index < 4) {
                        window.location.href = "/survey";
                    } else {
                        window.location.href = "/dashboard";
                    }
                }
            },
        });
    });
    if (window.location.pathname === "/dashboard") {
        var c = getCookie("session-id");
        if (c === "") return;
        var answersLabels = [
            ["Bored", "Excited", "Happy", "Sad"],
            ["Abandoned Farm", "Woods", "Busy city", "Sea"],
            ["Friends", "Family", "Strangers", "Authorities"],
            ["Friends", "Money", "Career Advancement", "Vacation"]
        ];
        $.ajax({
            url: '/api/aggregateResponses',
            type: 'post',
            dataType: 'json',
            data: {
            },
            success: function(data) {
                for (var i = 0; i < data.length; i++) {
                    var chartName = "chart" + (i+1).toString();
                    var chartTitle = "Question " + (i+1).toString();
                    var chartLabels = answersLabels[i]
                    var chartData = data[i];
                    initChart(chartName, chartTitle, chartLabels, chartData);
                }
            },
        });
    }
});

// initialize doughnut chart with user survey data
function initChart(id, chartTitle, chartLabels, chartData) {
    var ctx = $("#" + id);
    var chart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: chartLabels,
            datasets: [{
                label: '# of People',
                data: chartData,
                backgroundColor: [
                    'rgba(255, 99, 132, 0.2)',
                    'rgba(54, 162, 235, 0.2)',
                    'rgba(255, 206, 86, 0.2)',
                    'rgba(75, 192, 192, 0.2)',
                    'rgba(153, 102, 255, 0.2)',
                    'rgba(255, 159, 64, 0.2)'
                ],
                borderColor: [
                    'rgba(255,99,132,1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 206, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                    'rgba(153, 102, 255, 1)',
                    'rgba(255, 159, 64, 1)'
                ],
                borderWidth: 1
            }]
        },
        options: {
            title: {
                display: true,
                fontSize: 24,
                padding: 12,
                text: chartTitle
            }
        }
    });
}

// cookie functions from http://www.w3schools.com/js/js_cookies.asp
function setCookie(cname, cvalue, exdays) {
    var d = new Date();
    d.setTime(d.getTime() + (exdays*24*60*60*1000));
    var expires = "expires="+ d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function getCookie(cname) {
    var name = cname + "=";
    var ca = document.cookie.split(';');
    for(var i = 0; i <ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0)==' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length,c.length);
        }
    }
    return "";
}
