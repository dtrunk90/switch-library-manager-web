/*jshint esversion: 9 */
/*globals bootstrap */

function insertAlert(element, contextualClass, iconClass, strongMessage, message) {
	const alert = document.createElement('div');
	alert.classList.add('alert', contextualClass, 'd-flex', 'align-items-center', 'alert-dismissible', 'fade', 'show');
	alert.setAttribute('role', 'alert');

	const icon = document.createElement('span');
	icon.classList.add('bi', iconClass, 'flex-shrink-0', 'me-2');
	alert.appendChild(icon);

	const fullMessageWrapper = document.createElement('div');

	if (strongMessage) {
		const strongMessageWrapper = document.createElement('strong');
		strongMessageWrapper.appendChild(document.createTextNode(strongMessage));
		fullMessageWrapper.appendChild(strongMessageWrapper);
		fullMessageWrapper.appendChild(document.createTextNode(' '));
	}

	fullMessageWrapper.appendChild(document.createTextNode(message));
	alert.appendChild(fullMessageWrapper);

	const closeBtn = document.createElement('button');
	closeBtn.setAttribute('aria-label', 'Close');
	closeBtn.setAttribute('type', 'button');
	closeBtn.classList.add('btn-close');
	closeBtn.dataset.bsDismiss = 'alert';
	alert.appendChild(closeBtn);

	element.insertBefore(alert, element.firstChild);
}

function onSubmit(form) {
	const feedbackAlert = form.querySelector('.alert');
	if (feedbackAlert) {
		form.removeChild(feedbackAlert);
	}

	form.querySelectorAll('.is-invalid').forEach(e => e.classList.remove('is-invalid'));
	form.querySelectorAll('.invalid-feedback').forEach(e => e.parentNode.removeChild(e));

	const data = new FormData(form);
	fetch(form.action || window.location.href, {
		method: 'POST',
		body: new URLSearchParams(data).toString(),
		headers: {
			'Content-type': 'application/x-www-form-urlencoded'
		}
	}).then(function (response) {
		if (!response.ok) {
			throw response;
		}

		response.json().then(jsonResponse => {
			insertAlert(form, 'alert-success', 'bi-check-circle-fill', jsonResponse.strongMessage, jsonResponse.message);
		});
	}).catch(function (error) {
		error.json().then(jsonResponse => {
			if (jsonResponse.globalError.strongMessage || jsonResponse.globalError.message) {
				insertAlert(form, 'alert-danger', 'bi-exclamation-triangle-fill', jsonResponse.globalError.strongMessage, jsonResponse.globalError.message);
			} else if (jsonResponse.fieldErrors) {
				jsonResponse.fieldErrors.forEach(fieldError => {
					const validationFeedback = document.createElement('div');
					validationFeedback.id = `validation-feedback-${fieldError.field}`;
					validationFeedback.classList.add('invalid-feedback');
					validationFeedback.appendChild(document.createTextNode(fieldError.message));

					const field = form.querySelector(`[name="${fieldError.field}"]`);
					field.setAttribute('aria-describedby', `validation-feedback-${fieldError.field}`);
					field.classList.add('is-invalid');

					field.parentElement.appendChild(validationFeedback);
				});
			}
		});
	});
}

document.addEventListener('DOMContentLoaded', function () {
	const tooltipTriggerList = document.querySelectorAll('[data-bs-toggle="tooltip"]');
	[...tooltipTriggerList].map(tooltipTriggerEl => new bootstrap.Tooltip(tooltipTriggerEl));

	const settingsForm = document.getElementById('settingsForm');
	if (settingsForm) {
		settingsForm.addEventListener('submit', e => {
			e.preventDefault();
			onSubmit(settingsForm);
		});
	}
}, false);
