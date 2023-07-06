// Global application script will go here.

function loadIntoModal (method, url, data = {}) {
    $.ajax({
        url: url,
        type: method,
        data: jQuery.param(data) ,
        contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
    }).done(function (response) {
        console.log(response)
        showModalContent(response);
    }).fail(function (xhr, status, err) {
        showAjaxError("Error: "+ status);
        console.log(status)
        console.log(err)
        console.log(xhr)
    })
}

function showModalContent (response) {
    $('#modalTitle').html(response.modalTitle);
    $('#modalContent').html(response.modalContent);
    $('#modal').modal('show');    
}

function submitFormInBackground(formId) {
    let form = $(formId);
    if (form == null) {
        showAjaxAlert("Bad form id")
        return
    }
    //console.log(form.serialize())
    $.ajax({
        url: form.attr('action'),
        type: 'POST',
        data: form.serialize() ,
        contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
    }).done(function (response) {        
        //console.log(response)
    }).fail(function (xhr, status, err) {
        showAjaxError("Error: "+ status);
        console.log(status)
        console.log(err)
        console.log(xhr)
    })    
}

function showAjaxAlert(message) {
    $("#ajaxAlertMessage").html(message)
    $("#ajaxAlert").show()
}

function showAjaxError(message) {
    $("#ajaxErrorMessage").html(message)
    $("#ajaxError").show()
}
