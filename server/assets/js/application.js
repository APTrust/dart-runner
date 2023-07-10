// Global application script will go here.

function loadIntoModal (method, modalTitle, url, data = {}) {
    console.log(method)
    console.log(modalTitle)
    console.log(url)
    console.log(data)
    $.ajax({
        url: url,
        type: method,
        data: jQuery.param(data) ,
        contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
    }).done(function (response) {
        console.log(response)
        showModalContent(modalTitle, response);
    }).fail(function (xhr, status, err) {
        showAjaxError("Error: "+ status);
        console.log(status)
        console.log(err)
        console.log(xhr)
    })
}

function showModalContent (title, content) {
    $('#modalTitle').html(title);
    $('#modalContent').html(content);
    $('#modal').modal('show');    
}

function submitFormInBackground(formId) {
    console.log("submitFormInBackground")
    
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

function copyToClipboard(copySourceId, messageDivId) {
    let copyText = document.querySelector(copySourceId).textContent;
    navigator.clipboard.writeText(copyText)
    $(messageDivId).show();
    $(messageDivId).fadeOut({duration: 1800});
}