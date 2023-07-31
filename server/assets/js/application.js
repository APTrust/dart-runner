// Global JS

function loadIntoModal (method, modalTitle, url, data = {}) {
    $.ajax({
        url: url,
        type: method,
        data: jQuery.param(data) ,
        contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
    }).done(function (response) {
        //console.log(response)
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
    $('[data-toggle="popover"]').popover()
    $('#modal').modal('show');  
}

function submitFormInBackground(formId, successCallback, failureCallback) {
    console.log("submitFormInBackground")
    console.log(formId)
    
    let form = $(formId);
    if (form == null) {
        showAjaxAlert("Bad form id")
        return
    }
    console.log(form.serialize())
    $.ajax({
        url: form.attr('action'),
        type: 'POST',
        data: form.serialize() ,
        contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
    }).done(function (response) {        
        //console.log(response)
        if (successCallback != null) {
            successCallback(response)
        }
    }).fail(function (xhr, status, err) {
        if (failureCallback != null) {
            failureCallback(xhr, status, err)
        } else {
            showAjaxError("Error: "+ status)
        }
        console.log(status)
        console.log(err)
        console.log(xhr)
    })    
}

function showAjaxAlert(message) {
    console.log("Ajax Alert: " + message)
    $("#ajaxAlertMessage").html(message)
    $("#ajaxAlert").show()
}

function showAjaxError(message) {
    console.log("Ajax Error: " + message)
    $("#ajaxErrorMessage").html(message)
    $("#ajaxError").show()
}

function copyToClipboard(copySourceId, messageDivId) {
    let copyText = document.querySelector(copySourceId).textContent;
    navigator.clipboard.writeText(copyText)
    $(messageDivId).show();
    $(messageDivId).fadeOut({duration: 1800});
}

function submitTagDefForm(formId) {
    let onSuccess = function(response) {
        // We don't need to do anything here, since a successful
        // save results in a redirect.
        console.log("Tag definition operation succeeded")
        console.log(response)
        location.href = response.location
    }
    let onFail = function(xhr, status, err) {
        // Failure is typically a validation failure.
        // Re-display the form to show specific error messages.
        let modalTitle = $('#modalTitle').html()
        showModalContent(modalTitle, xhr.responseText)        
    }
    submitFormInBackground(formId, onSuccess, onFail)
}

function submitNewTagFileForm(formId) {
    let onSuccess = function(response) {
        // We don't need to do anything here, since a successful
        // save results in a redirect.
        console.log("Created new tag file")
        console.log(response)
        location.href = response.location
    }
    let onFail = function(xhr, status, err) {
        // Failure is typically a validation failure.
        // Re-display the form to show specific error messages.
        let modalTitle = $('#modalTitle').html()
        showModalContent(modalTitle, xhr.responseText)        
    }
    submitFormInBackground(formId, onSuccess, onFail)
}

function confirmBackgroundDeletion(question, url, data) {
    if (confirm(question)) {
        $.ajax({
            url: url,
            type: "post",
            data: data ? jQuery.param(data) : null,
            contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
        }).done(function (response) {
            location.href = response.location
        }).fail(function (xhr, status, err) {
            showModalContent("Error deleting job", xhr.responseText)
        })    
    }
}

function deleteProfile(formId) {
    if (confirm("Delete this BagIt profile?")) {
        submitTagDefForm(formId)
    }
}

function openExternalUrl(url) {
    let data = { "url": url }
    $.ajax({
        url: "/open_external",
        type: "get",
        data: jQuery.param(data),
    }).done(function (response) {
        console.log("Show help succeeded")
    }).fail(function (xhr, status, err) {
        alert(xhr.responseText)        
    })        
}

function execCmd(url) {
    $.ajax({
        url: url,
        type: "get",
    }).done(function (response) {
        console.log("Exec command succeeded")
    }).fail(function (xhr, status, err) {
        alert(xhr.responseText)        
    })        
}

// Page init
$(function () {
    $('[data-toggle="popover"]').popover()
})
