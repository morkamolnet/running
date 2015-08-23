var distance = 1200;
var days = 365;
var daily = 3.29;
var arr = [];
var locale = "en-us";
var d = new Date("1/1/2015");
var today = new Date();
var oneDay = 24*60*60*1000;
var diffdays = Math.round(Math.abs((d.getTime() - today.getTime())/(oneDay)));

var dates = []
currentmonth = 0;
for(i=0;i<days;i++){
    if(arr.length > 0){
        arr.push(arr[arr.length-1]+daily)
        if(currentmonth != d.getMonth()){
            dates.push(d.toLocaleString(locale, { month: 'short' }));
            currentmonth = d.getMonth();
        }
        else{
            dates.push('')
        }
    } 
    else{
        arr.push(daily)
        dates.push('Jan')
    }
    d.setDate(d.getDate() + 1);
}


$.getJSON( "/getstats/", function( result ) {
  current = result.Thisyear;
  prev = result.Lastyear;

  distdiff = parseInt(arr[diffdays] - current[diffdays]);


  var data = {
    labels : dates,
    datasets : [
        {
            label: "This year",
            fillColor: "rgba(151,187,205,0.2)",
            strokeColor: "rgba(151,187,205,1)",
            pointColor: "rgba(151,187,205,1)",
            pointStrokeColor: "#fff",
            pointHighlightFill: "#fff",
            pointHighlightStroke: "rgba(151,187,205,1)",
            data : current
        },

        {
            label: "Prev year",
            fillColor: "rgba(151,187,205,0.0)",
            strokeColor: "rgba(255,255,153,0.3)",
            pointColor: "rgba(151,187,205,1)",
            pointStrokeColor: "#fff",
            pointHighlightFill: "#fff",
            pointHighlightStroke: "rgba(151,187,205,1)",
            data : prev
        },

        {
            label: "Target",
            fillColor: "rgba(220,220,220,0.0)",
            strokeColor: "rgba(253,159,52,0.9)",
            pointColor: "rgba(220,220,220,1)",
            pointStrokeColor: "#fff",
            pointHighlightFill: "#fff",
            pointHighlightStroke: "rgba(220,220,220,1)",
            data: arr
        }

        ]
    }

    var ctx = document.getElementById("LineWithLine").getContext("2d");

    Chart.types.Line.extend({
        name: "LineWithLine",
        initialize: function () {
            Chart.types.Line.prototype.initialize.apply(this, arguments);
        },
        draw: function () {
            Chart.types.Line.prototype.draw.apply(this, arguments);
            
            var point = this.datasets[0].points[this.options.lineAtIndex]
            var scale = this.scale

            // draw line
            this.chart.ctx.beginPath();
            this.chart.ctx.moveTo(point.x, scale.startPoint + 24);
            this.chart.ctx.strokeStyle = '#c0392b';
            this.chart.ctx.lineTo(point.x, scale.endPoint);
            this.chart.ctx.stroke();
            this.chart.ctx.fontSize + 23;
            
            // write TODAY
            this.chart.ctx.textAlign = 'center';
            this.chart.ctx.fillText("TODAY, diff "+ distdiff+ "km", point.x, scale.startPoint + 12);
        }
    });



    var myNewChart = new Chart(ctx).LineWithLine(data, {
        scaleShowLabels: true,
        pointDot : false,
        showTooltips: false,
        animation: false,
        scaleFontColor: "#ccc",
        bezierCurve : false,
        scaleOverride: true,
        scaleSteps: 12,

        // Number - The value jump in the hard coded scale
        scaleStepWidth: 100,
        // Number - The scale starting value
        scaleStartValue: 0,
        scaleShowGridLines : false,
        datasetFill : false,
        lineAtIndex: diffdays

    });
});