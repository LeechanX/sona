<!DOCTYPE html>
<html>
<head>
    <title>Sona</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <link rel="stylesheet" href="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/css/bootstrap.min.css">
    <link rel="stylesheet" href="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/css/bootstrap-theme.min.css">
    {{.HtmlHead}}
</head>
<body>

    <div class="container">
        <h3>{{.ServiceKey}} <small>Version: {{.Version}}</small></h3>
        <p class="text-danger">Submit才会最终执行操作, 故操作完成后请点击Submit</p>
        <br>
    <form id="edit_form" class="form-horizontal" action="/run/update" method="post" onsubmit="return option_all()">
        <select multiple class="form-control" id="conf_data_id" name="confs">
            {{range $key, $value := .Confs}}
            <option disabled="disabled" value="{{$key}} = {{$value}}">{{$key}} = {{$value}}</option>
            {{end}}
        </select>
        <input type="hidden" name="servicekey" value="{{.ServiceKey}}">
        <input type="hidden" name="version" value="{{.Version}}">
        <br>
        <label for="exampleInputName2">Target Key</label>
        <input type="text" class="form-control" id="target_key_id" placeholder="section.configureKey">
        <br>
        <label for="exampleInputName2">Value</label>
        <input type="text" class="form-control" id="target_value_id" placeholder="">

        <br/>

        <button type="button" class="btn btn-info" id="update_btn_id">add/update one</button>
        <button type="button" class="btn btn-warning" id="delete_btn_id">delete</button>
        <button type="button" class="btn btn-danger" id="clear_btn_id">clear</button>
        <br/>
        <br/>
        <div id="operation_log">
        </div>
        <br/>
        <div class="form-group">
        <div class="col-sm-10">
          <button type="submit" class="btn btn-primary btn-lg">Submit</button>
        </div>
      </div>
        <br/>
    </form>

    <blockquote id="result">
    </blockquote>

    <script type="text/javascript" src="http://code.jquery.com/jquery-2.0.3.min.js"></script>
    <script src="http://netdna.bootstrapcdn.com/bootstrap/3.0.3/js/bootstrap.min.js"></script>
    <script src="http://malsup.github.com/jquery.form.js"></script> 
    <script type="text/javascript">
        function option_all() {
            $('#conf_data_id option').each(function () {
                $(this).attr('selected', true); 
            });
            //var values = $("#conf_data_id>option").map(function() { return $(this).val(); });
            //$('#conf_data_id').val(values)
        }
        $('#update_btn_id').click(function() {
            var key = $('#target_key_id').val().trim()
            var value = $('#target_value_id').val().trim()
            if (key == "") {
                alert("configureKey is empty")
                return
            }
            if (value == "") {
                alert("value is empty")
                return
            }
            if (value.length > 200) {
                alert("value's length shouldn't exceed 200 bytes")
                return
            }
            var attrs = key.split(".")
            if (attrs.length != 2) {
                alert('key format error')
                return
            }
            if (attrs[0].length <= 0) {
                alert("section key is empty")
                return
            }
            if (attrs[0].length <= 0 || attrs[0].length > 30) {
                alert("section's length shouldn't exceed 30 bytes")
                return
            }
            if (attrs[1].length <= 0) {
                alert("configure key is empty")
                return
            }
            if (attrs[1].length > 30) {
                alert("configure key's length shouldn't exceed 30 bytes")
                return
            }
            var exist = false
            $("#conf_data_id option").each(function() {
                var originVal = $(this).val();
                if (originVal.split(" = ")[0] == key) {
                    exist = true
                    //delete it
                    $(this).remove()
                    return false;
                }
            })

            var optionValue = key + " = " + value
            $("#conf_data_id").append("<option disabled=\"disabled\" value='"+ optionValue +"'>" + optionValue + "</option>")

            var log = $("#operation_log").html();
            var result = "added"
            if (exist) {
                result = "updated"
            }
            $("#operation_log").html(log + "<p class=\"text-primary\">" + key + " is " + result + "</p>")

        });
        $('#delete_btn_id').click(function(){
            var key = $('#target_key_id').val().trim();
            if (key == "") {
                alert("configureKey is empty")
                return
            }
            var exist = false
            $("#conf_data_id option").each(function() {
                var originVal = $(this).val();
                if (originVal.split(" = ")[0] == key) {
                    exist = true
                    //delete it
                    $(this).remove()
                    return false;
                }
            })

            if (exist == false) {
                alert("no such key")
            } else {
                var log = $("#operation_log").html();
                $("#operation_log").html(log + "<p class=\"text-warning\">" + key + " is deleted" + "</p>")
            }
        });
        $('#clear_btn_id').click(function(){
            $("#conf_data_id").empty(); 
            var log = $("#operation_log").html();
            $("#operation_log").html(log + "<p class=\"text-danger\">all configures are cleared</p>")
        });

        $("#edit_form").ajaxForm(function(data) {
        msg = eval(data) //转化为json数据
        if(msg.code == 200) {
          $("#result").html("<p class=\"text-success\">update successfully</p>")
        } else {
          $("#result").html("<p class=\"text-danger\"> update error: " + msg.message + "</p>")
        }
      });
    </script>

</body>
</html>
