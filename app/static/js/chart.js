google.charts.load('current', {
    'packages': ['corechart', 'controls']
});

var config = {
    api: {
        enable: true,
        interval: 1000 * 3
    },
    candlestick: {
        product_code: 'BTC_JPY',
        duration: '5m',
        limit: 365,
        numViews: 5,
    },
    dataTable: {
        index: 0,
        value: null
    },
    sma: {
        enable: false,
        indexes: [],
        periods: [],
        values: []
    },
    ema: {
        enable: false,
        indexes: [],
        periods: [],
        values: []
    },
    bbands: {
        enable: false,
        indexes: [],
        n: 20,
        k: 2,
        up: [],
        mid: [],
        down: []
    },
    ichimoku: {
        enable: false,
        indexes: [],
        tenkan: [],
        kijun: [],
        senkouA: [],
        senkouB: [],
        chikou: []
    },
    volume: {
        enable: false,
        index: [],
        values: []
    },
    rsi: {
        enable: false,
        indexes: {
            'up': 0,
            'value': 0,
            'down': 0
        },
        period: 14,
        up: 70,
        values: [],
        down: 30
    },
    macd: {
        enable: false,
        indexes: [],
        periods: [],
        values: []
    },
    hv: {
        enable: false,
        indexes: [],
        periods: [],
        values: []
    },
    events: {
        enable: false,
        indexes: [],
        values: [],
        first: null
    }
};

function initConfigValues() {
    config.dataTable.index = 0;
    config.sma.indexes = [];
    config.sma.values = [];
    config.ema.indexes = [];
    config.ema.values = [];
    config.bbands.indexes = [];
    config.bbands.up = [];
    config.bbands.mid = [];
    config.bbands.down = [];
    config.ichimoku.indexes = [];
    config.ichimoku.tenkan = [];
    config.ichimoku.kijun = [];
    config.ichimoku.senkouA = [];
    config.ichimoku.senkouB = [];
    config.ichimoku.chikou = [];
    config.volume.index = [];
    config.rsi.indexes = [];
    config.macd.indexes = [];
    config.macd.periods = [];
    config.macd.values = [];
    config.hv.indexes = [];
    config.hv.periods = [];
    config.hv.values = [];
    config.events.indexes = [];
    config.events.values = [];
}

// candlestick chart を描画する処理
// send() 関数の中で呼び出される
function drawChart(dataTable) {
    var chartDiv = document.getElementById('chart_div');
    var charts = [];
    var dashboard = new google.visualization.Dashboard(chartDiv);
    // 描画するグラフに関する設定を記述。ここではコンボチャートを作成する。
    var mainChart = new google.visualization.ChartWrapper({
        chartType: 'ComboChart',
        containerId: 'chart_div',
        options: {
            backgroundColor: "black",
            // 横軸に関する設定
            hAxis: {
                // trueの場合、水平軸のテキストを斜めに描画して、軸に沿ってより多くのテキストを収めるようにする
                'slantedText': true
            },
            legend: {
                // 凡例の位置
                'position': 'none'
            },
            candlestick: {
                fallingColor: {
                    strokeWidth: 0,
                    fill: '#a52714'
                },
                risingColor: {
                    strokeWidth: 0,
                    fill: '#0f9d58'
                }
            },
            seriesType: "candlesticks",
            series: {}
        },
        view: {
            columns: [{
                // function を設定
                calc: function (d, rowIndex) {
                    return d.getFormattedValue(rowIndex, 0);
                },
                type: 'string'

            }, 1, 2, 3, 4]

        }

    });
    // chart 配列に Chartwrapper インスタンスを追加
    charts.push(mainChart);

    // インスタンス mainChart の option 情報を格納
    var options = mainChart.getOptions();
    // view = {columns: [{calc: f, type: "string"}, 1, 2, 3, 4]}
    var view = mainChart.getView();

    if (config.sma.enable == true) {
        for (i = 0; i < config.sma.indexes.length; i++) {
            options.series[config.sma.indexes[i]] = {
                type: 'line'
            };
            view.columns.push(config.candlestick.numViews + config.sma.indexes[i]);
        }
    }

    if (config.ema.enable == true) {
        for (i = 0; i < config.ema.indexes.length; i++) {
            options.series[config.ema.indexes[i]] = {
                type: 'line'
            };
            view.columns.push(config.candlestick.numViews + config.ema.indexes[i]);
        }
    }

    if (config.bbands.enable == true) {
        for (i = 0; i < config.bbands.indexes.length; i++) {
            options.series[config.bbands.indexes[i]] = {
                type: 'line',
                color: 'blue',
                lineWidth: 1
            };
            view.columns.push(config.candlestick.numViews + config.bbands.indexes[i])
        }
    }

    if (config.ichimoku.enable == true) {
        for (i = 0; i < config.ichimoku.indexes.length; i++) {
            options.series[config.ichimoku.indexes[i]] = {
                type: 'line',
                lineWidth: 1
            };
            view.columns.push(config.candlestick.numViews + config.ichimoku.indexes[i]);
        }
    }

    if (config.events.enable == true && config.events.indexes.length > 0) {
        options.series[config.events.indexes[0]] = {
            'type': 'line',
            tooltip: 'none',
            enableInteractivity: false,
            lineWidth: 0
        };
        view.columns.push(config.candlestick.numViews + config.events.indexes[0]);
        view.columns.push(config.candlestick.numViews + config.events.indexes[1]);
    }

    if (config.volume.enable == true) {
        if ($('#volume_div').length == 0) {
            $('#technical_div').append(
                "<div id='volume_div' class='bottom_chart'>" +
                "<span class='technical_title'>Volume</span>" +
                "<div id='volume_chart'></div>" +
                "</div>")
        }
        var volumeChart = new google.visualization.ChartWrapper({
            'chartType': 'ColumnChart',
            'containerId': 'volume_chart',
            'options': {
                'backgroundColor': "black",
                'hAxis': {
                    'slantedText': false
                },
                'legend': {
                    'position': 'none'
                },
                'series': {}
            },
            'view': {
                'columns': [{
                    'type': 'string'
                }, 5]
            }
        });
        charts.push(volumeChart)
    }

    if (config.rsi.enable == true) {
        if ($('#rsi_div').length == 0) {
            $('#technical_div').append(
                "<div id='rsi_div' class='bottom_chart'>" +
                "<span class='technical_title'>RSI</span>" +
                "<div id='rsi_chart'></div>" +
                "</div>")
        }
        var up = config.candlestick.numViews + config.rsi.indexes['up'];
        var value = config.candlestick.numViews + config.rsi.indexes['value'];
        var down = config.candlestick.numViews + config.rsi.indexes['down'];
        var rsiChart = new google.visualization.ChartWrapper({
            'chartType': 'LineChart',
            'containerId': 'rsi_chart',
            'options': {
                'backgroundColor': "black",
                'hAxis': {
                    'slantedText': false
                },
                'legend': {
                    'position': 'none'
                },
                'series': {
                    0: {
                        color: 'black',
                        lineWidth: 1
                    },
                    1: {
                        color: '#e2431e',
                    },
                    2: {
                        color: 'black',
                        lineWidth: 1
                    }
                }
            },
            'view': {
                'columns': [{
                    'type': 'string'
                }, up, value, down]
            }
        });
        charts.push(rsiChart)
    }

    if (config.macd.enable == true) {
        if (config.macd.indexes.length == 0) {
            return
        }
        if ($('#macd_div').length == 0) {
            $('#technical_div').append(
                "<div id='macd_div'>" +
                "<span class='technical_title'>MACD</span>" +
                "<div id='macd_chart'></div>" +
                "</div>")
        }
        var macdColumns = [{
            'type': 'string'
        }];

        macdColumns.push(config.candlestick.numViews + config.macd.indexes[2]);
        macdColumns.push(config.candlestick.numViews + config.macd.indexes[0]);
        macdColumns.push(config.candlestick.numViews + config.macd.indexes[1]);
        var macdChart = new google.visualization.ChartWrapper({
            'chartType': 'ComboChart',
            'containerId': 'macd_chart',
            'options': {
                backgroundColor: "black",
                legend: {
                    'position': 'none'
                },
                seriesType: "bars",
                series: {
                    1: {
                        type: 'line',
                        lineWidth: 1
                    },
                    2: {
                        type: 'line',
                        lineWidth: 1
                    }
                }
            },
            'view': {
                'columns': macdColumns
            }
        });
        charts.push(macdChart)
    }

    if (config.hv.enable == true) {
        if (config.hv.indexes.length == 0) {
            return
        }
        if ($('#hv_div').length == 0) {
            $('#technical_div').append(
                "<div id='hv_div'>" +
                "<span class='technical_title'>Hv</span>" +
                "<div id='hv_chart'></div>" +
                "</div>")
        }
        var hvSeries = {};
        var hvColumns = [{
            'type': 'string'
        }];

        for (i = 0; i < config.hv.indexes.length; i++) {
            hvSeries[config.hv.indexes[i]] = {
                lineWidth: 1
            };
            hvColumns.push(config.candlestick.numViews + config.hv.indexes[i]);
        }
        var hvChart = new google.visualization.ChartWrapper({
            'chartType': 'LineChart',
            'containerId': 'hv_chart',
            'options': {
                'backgroundColor': "black",
                'legend': {
                    'position': 'none'
                },
                'series': hvSeries
            },
            'view': {
                'columns': hvColumns
            }
        });
        charts.push(hvChart)
    }

    // candlestick グラフのフィルター制御を行うインスタンスの定義
    var controlWrapper = new google.visualization.ControlWrapper({
        'controlType': 'ChartRangeFilter',
        'containerId': 'filter_div',
        'options': {
            'backgroundColor': "black",
            'filterColumnIndex': 0,
            'ui': {
                'chartType': 'LineChart',
                'chartView': {
                    'columns': [0, 4]
                },
                'chartOptions': {
                    'backgroundColor': "black"
                }
            }
        }
    });

    dashboard.bind(controlWrapper, charts);
    dashboard.draw(dataTable);

}

// setInterval() 関数によって定期的に実行される
function send() {
    if (config.api.enable == false) {
        return
    }
    var params = {
        "product_code": config.candlestick.product_code,
        "limit": config.candlestick.limit,
        "duration": config.candlestick.duration,
    }

    if (config.sma.enable == true) {
        params["sma"] = true;
        params["smaPeriod1"] = config.sma.periods[0];
        params["smaPeriod2"] = config.sma.periods[1];
        params["smaPeriod3"] = config.sma.periods[2];
    }

    if (config.ema.enable == true) {
        params["ema"] = true;
        params["emaPeriod1"] = config.ema.periods[0];
        params["emaPeriod2"] = config.ema.periods[1];
        params["emaPeriod3"] = config.ema.periods[2];
    }

    if (config.bbands.enable == true) {
        params["bbands"] = true;
        params["bbandsN"] = config.bbands.n;
        params["bbandsK"] = config.bbands.k;
    }

    if (config.ichimoku.enable == true) {
        params["ichimoku"] = true;
    }

    if (config.rsi.enable == true) {
        params["rsi"] = true;
        params["rsiPeriod"] = config.rsi.period;
    }

    if (config.macd.enable == true) {
        params["macd"] = true;
        params["macdPeriod1"] = config.macd.periods[0];
        params["macdPeriod2"] = config.macd.periods[1];
        params["macdPeriod3"] = config.macd.periods[2];
    }

    if (config.hv.enable == true) {
        params["hv"] = true;
        params["hvPeriod1"] = config.hv.periods[0];
        params["hvPeriod2"] = config.hv.periods[1];
        params["hvPeriod3"] = config.hv.periods[2];
    }

    if (config.events.enable == true) {
        params["events"] = true;
    }

    $.get("/api/candle/", params).done(function (data) {
        initConfigValues();
        aidata = data["ai"];
        $("#description").text(JSON.stringify(aidata.OptimizedTradeParams));
        data = data["dfcandle"];
        var dataTable = new google.visualization.DataTable();
        dataTable.addColumn('datetime', 'Date');
        dataTable.addColumn('number', 'Low');
        dataTable.addColumn('number', 'Open');
        dataTable.addColumn('number', 'Close');
        dataTable.addColumn('number', 'High');
        dataTable.addColumn('number', 'Volume');

        if (data["smas"] != undefined) {
            for (i = 0; i < data['smas'].length; i++) {
                var smaData = data['smas'][i];
                if (smaData.length == 0) {
                    continue;
                }
                config.dataTable.index += 1;
                config.sma.indexes[i] = config.dataTable.index;
                dataTable.addColumn('number', 'SMA' + smaData["period"].toString());
                config.sma.values[i] = smaData["values"]
            }
        }

        if (data["emas"] != undefined) {
            for (i = 0; i < data['emas'].length; i++) {
                var emaData = data['emas'][i];
                if (emaData.length == 0) {
                    continue;
                }
                config.dataTable.index += 1;
                config.ema.indexes[i] = config.dataTable.index;
                dataTable.addColumn('number', 'EMA' + emaData["period"].toString());
                config.ema.values[i] = emaData["values"]
            }
        }

        if (data['bbands'] != undefined) {
            var n = data['bbands']['n'];
            var k = data['bbands']['k'];
            var up = data['bbands']['up'];
            var mid = data['bbands']['mid'];
            var down = data['bbands']['down'];
            config.dataTable.index += 1;
            config.bbands.indexes[0] = config.dataTable.index;
            config.dataTable.index += 1;
            config.bbands.indexes[1] = config.dataTable.index;
            config.dataTable.index += 1;
            config.bbands.indexes[2] = config.dataTable.index;
            dataTable.addColumn('number', 'BBands Up(' + n + ',' + k + ')');
            dataTable.addColumn('number', 'BBands Mid(' + n + ',' + k + ')');
            dataTable.addColumn('number', 'BBands Down(' + n + ',' + k + ')');
            config.bbands.up = up;
            config.bbands.mid = mid;
            config.bbands.down = down;
        }

        if (data['ichimoku'] != undefined) {
            var tenkan = data['ichimoku']['tenkan'];
            var kijun = data['ichimoku']['kijun'];
            var senkouA = data['ichimoku']['senkoua'];
            var senkouB = data['ichimoku']['senkoub'];
            var chikou = data['ichimoku']['chikou'];

            config.dataTable.index += 1;
            config.ichimoku.indexes[0] = config.dataTable.index;
            config.dataTable.index += 1;
            config.ichimoku.indexes[1] = config.dataTable.index;
            config.dataTable.index += 1;
            config.ichimoku.indexes[2] = config.dataTable.index;
            config.dataTable.index += 1;
            config.ichimoku.indexes[3] = config.dataTable.index;
            config.dataTable.index += 1;
            config.ichimoku.indexes[4] = config.dataTable.index;

            config.ichimoku.tenkan = tenkan;
            config.ichimoku.kijun = kijun;
            config.ichimoku.senkouA = senkouA;
            config.ichimoku.senkouB = senkouB;
            config.ichimoku.chikou = chikou;

            dataTable.addColumn('number', 'Tenkan');
            dataTable.addColumn('number', 'Kijun');
            dataTable.addColumn('number', 'SenkouA');
            dataTable.addColumn('number', 'SenkouB');
            dataTable.addColumn('number', 'Chikou');
        }


        if (data['events'] != undefined) {
            config.dataTable.index += 1;
            config.events.indexes[0] = config.dataTable.index;
            config.dataTable.index += 1;
            config.events.indexes[1] = config.dataTable.index;

            config.events.values = data['events']['signals'];
            config.events.first = config.events.values.shift();

            dataTable.addColumn('number', 'Marker');
            dataTable.addColumn({
                type: 'string',
                role: 'annotation'
            });


            if (data['events']['profit'] != undefined) {
                profit = "$" + String(Math.round(data['events']['profit'] * 100) / 100);
                $('#profit').html("Change:" + profit);
            }
        }

        if (data['rsi'] != undefined) {
            console.log(data);
            config.dataTable.index += 1;
            config.rsi.indexes['up'] = config.dataTable.index;
            config.dataTable.index += 1;
            config.rsi.indexes['value'] = config.dataTable.index;
            config.dataTable.index += 1;
            config.rsi.indexes['down'] = config.dataTable.index;
            config.rsi.period = data['rsi']['period'];
            config.rsi.values = data['rsi']['values'];
            dataTable.addColumn('number', 'RSI Thread');
            dataTable.addColumn('number', 'RSI(' + config.rsi.period + ')');
            dataTable.addColumn('number', 'RSI Thread');
        }

        if (data['macd'] != undefined) {
            var macdData = data['macd'];
            var fast_period = macdData["fast_period"].toString();
            var slow_period = macdData["slow_period"].toString();
            var signal_period = macdData["signal_period"].toString();
            var macd = macdData["macd"];
            var macd_signal = macdData["macd_signal"];
            var macd_hist = macdData["macd_hist"];

            config.dataTable.index += 1;
            config.macd.indexes[0] = config.dataTable.index;
            config.dataTable.index += 1;
            config.macd.indexes[1] = config.dataTable.index;
            config.dataTable.index += 1;
            config.macd.indexes[2] = config.dataTable.index;
            var speriods = '(' + fast_period + ',' + slow_period + ',' + signal_period + ')';
            dataTable.addColumn('number', 'MD' + speriods);
            dataTable.addColumn('number', 'MS' + speriods);
            dataTable.addColumn('number', 'HT' + speriods);
            config.macd.values[0] = macd;
            config.macd.values[1] = macd_signal;
            config.macd.values[2] = macd_hist;
            config.macd.periods[0] = fast_period;
            config.macd.periods[1] = slow_period;
            config.macd.periods[2] = signal_period;
        }

        if (data['hvs'] != undefined) {
            for (i = 0; i < data['hvs'].length; i++) {
                var hvData = data['hvs'][i];
                if (hvData.length == 0) {
                    continue;
                }

                var period = hvData["period"].toString();
                var value = hvData["values"];

                config.dataTable.index += 1;
                config.hv.indexes[i] = config.dataTable.index;

                dataTable.addColumn('number', 'HV(' + period + ')');
                config.hv.values[i] = hvData["values"];
                config.hv.periods[i] = period;
            }
        }

        var googleChartData = [];
        var candles = data["candles"];


        for (var i = 0; i < candles.length; i++) {
            var candle = candles[i];
            var date = new Date(candle.time);
            var datas = [date, candle.low, candle.open, candle.close, candle.high, candle.volume];

            if (data["smas"] != undefined) {
                for (j = 0; j < config.sma.values.length; j++) {
                    if (config.sma.values[j][i] == 0) {
                        datas.push(null);
                    } else {
                        datas.push(config.sma.values[j][i]);
                    }
                }
            }

            if (data["emas"] != undefined) {
                for (j = 0; j < config.ema.values.length; j++) {
                    if (config.ema.values[j][i] == 0) {
                        datas.push(null);
                    } else {
                        datas.push(config.ema.values[j][i]);
                    }
                }
            }

            if (data["bbands"] != undefined) {
                if (config.bbands.up[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.bbands.up[i]);
                }
                if (config.bbands.mid[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.bbands.mid[i]);
                }
                if (config.bbands.down[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.bbands.down[i]);
                }
            }

            if (data["ichimoku"] != undefined) {
                if (config.ichimoku.tenkan[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.ichimoku.tenkan[i]);
                }
                if (config.ichimoku.kijun[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.ichimoku.kijun[i]);
                }
                if (config.ichimoku.senkouA[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.ichimoku.senkouA[i]);
                }
                if (config.ichimoku.senkouB[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.ichimoku.senkouB[i]);
                }
                if (config.ichimoku.chikou[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.ichimoku.chikou[i]);
                }
            }

            if (data['events'] != undefined) {
                var event = config.events.first
                if (event == undefined) {
                    datas.push(null);
                    datas.push(null);
                // } else if (event.time == candle.time) {
                } else if (floorDatetime(event.time, aidata.Duration).toString() == floorDatetime(candle.time, aidata.Duration).toString()) {
                    datas.push(candle.high + 1);
                    datas.push(event.side);
                    // config.events.first = config.events.values.shift();

                    do {
                        config.events.first = config.events.values.shift();
                        // firstEvent = config.events.first;
                    // } while (config.events.first.time == event.time);
                    // } while (config.events.first != undefined && config.events.first.time == event.time);
                    } while (config.events.first != undefined && floorDatetime(config.events.first.time, aidata.Duration).toString() == floorDatetime(event.time, aidata.Duration).toString());
                } else {
                    datas.push(null);
                    datas.push(null);
                }
            }

            if (data["rsi"] != undefined) {
                datas.push(config.rsi.up);
                if (config.rsi.values[i] == 0) {
                    datas.push(null);
                } else {
                    datas.push(config.rsi.values[i]);
                }
                datas.push(config.rsi.down);
            }

            if (data["macd"] != undefined) {
                for (j = 0; j < config.macd.values.length; j++) {
                    if (config.macd.values[j][i] == 0) {
                        datas.push(null);
                    } else {
                        datas.push(config.macd.values[j][i]);
                    }
                }
            }

            if (data["hvs"] != undefined) {
                for (j = 0; j < config.hv.values.length; j++) {
                    if (config.hv.values[j][i] == 0) {
                        datas.push(null);
                    } else {
                        datas.push(config.hv.values[j][i]);
                    }
                }
            }

            googleChartData.push(datas)
        }

        dataTable.addRows(googleChartData);
        var formatter_long = new google.visualization.DateFormat({formatType: 'MM/dd\nHH:mm:ss'});
        formatter_long.format(dataTable, 0);
        drawChart(dataTable);
    })
}

function changeDuration(s) {
    config.candlestick.duration = s;
    send();
}

function floorDatetime(date_time, duration) {
    var date = new Date(date_time);  
    var interval = duration / 60000000000;
    var coeff = 1000 * 60 * interval;
    var rounded_date = new Date(Math.floor(date.getTime() / coeff) * coeff);
    return rounded_date
}

setInterval(send, 1000 * 10)
// 全ての DOM ツリー構造および関連リソースが読み込まれたタイミングで実行される
window.onload = function () {
    send()

    // dashboard 部分にマウスがある場合は send関数 を実行しない(グラフの更新を行わない)
    $('#dashboard_div').mouseenter(function () {
        config.api.enable = false;
    }).mouseleave(function () {
        config.api.enable = true;
    });

    $('#inputSma').change(function () {
        if (this.checked === true) {
            config.sma.enable = true;
        } else {
            config.sma.enable = false;
        }
        send();
    });
    $("#inputSmaPeriod1").change(function () {
        config.sma.periods[0] = this.value;
        send();
    });
    $("#inputSmaPeriod2").change(function () {
        config.sma.periods[1] = this.value;
        send();
    });
    $("#inputSmaPeriod3").change(function () {
        config.sma.periods[2] = this.value;
        send();
    });

    $('#inputEma').change(function () {
        if (this.checked === true) {
            config.ema.enable = true;
        } else {
            config.ema.enable = false;
        }
        send();
    });
    $("#inputEmaPeriod1").change(function () {
        config.ema.periods[0] = this.value;
        send();
    });
    $("#inputEmaPeriod2").change(function () {
        config.ema.periods[1] = this.value;
        send();
    });
    $("#inputEmaPeriod3").change(function () {
        config.ema.periods[2] = this.value;
        send();
    });

    $('#inputBBands').change(function () {
        if (this.checked === true) {
            config.bbands.enable = true;
        } else {
            config.bbands.enable = false;
        }
        send();
    });
    $("#inputBBandsN").change(function () {
        config.bbands.n = this.value;
        send();
    });
    $("#inputBBandsK").change(function () {
        config.bbands.k = this.value;
        send();
    });

    $('#inputIchimoku').change(function () {
        if (this.checked === true) {
            config.ichimoku.enable = true;
        } else {
            config.ichimoku.enable = false;
        }
        send();
    });

    $('#inputVolume').change(function () {
        if (this.checked === true) {
            config.volume.enable = true;
            drawChart(config.dataTable.value);
        } else {
            config.volume.enable = false;
            $('#volume_div').remove();
        }
    });

    $('#inputRsi').change(function () {
        if (this.checked === true) {
            console.log("inputRsi")
            config.rsi.enable = true;
        } else {
            config.rsi.enable = false;
            $('#rsi_div').remove();
        }
        send();
    });
    $("#inputRsiPeriod").change(function () {
        config.rsi.period = this.value;
        send();
    });

    $('#inputMacd').change(function () {
        if (this.checked === true) {
            config.macd.enable = true;
        } else {
            $('#macd_div').remove();
            config.macd.enable = false;
        }
        send();
    });
    $("#inputMacdPeriod1").change(function () {
        config.macd.periods[0] = this.value;
        send();
    });
    $("#inputMacdPeriod2").change(function () {
        config.macd.periods[1] = this.value;
        send();
    });
    $("#inputMacdPeriod3").change(function () {
        config.macd.periods[2] = this.value;
        send();
    });

    $('#inputHv').change(function () {
        if (this.checked === true) {
            config.hv.enable = true;
        } else {
            $('#hv_div').remove();
            config.hv.enable = false;
        }
        send();
    });
    $("#inputHvPeriod1").change(function () {
        config.hv.periods[0] = this.value;
        send();
    });
    $("#inputHvPeriod2").change(function () {
        config.hv.periods[1] = this.value;
        send();
    });
    $("#inputHvPeriod3").change(function () {
        config.hv.periods[2] = this.value;
        send();
    });

    $('#inputEvents').change(function () {
        if (this.checked === true) {
            config.events.enable = true;
        } else {
            config.events.enable = false;
            $('#profit').html("");
        }
        send();
    });

}