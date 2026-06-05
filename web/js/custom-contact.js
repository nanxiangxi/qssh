(function($) {
	$("#contact_form").validate({
		submitHandler: function(form) {
			var form_btn = $(form).find('button[type="submit"]');
			var form_result_div = '#form-result';
			$(form_result_div).remove();
			form_btn.before('<div id="form-result" class="alert alert-success" role="alert" style="display: none;"></div>');
			var form_btn_old_msg = form_btn.html();
			form_btn.html(form_btn.prop('disabled', true).data("loading-text"));
			$(form).ajaxSubmit({
				dataType:  'json',
				success: function(data) {
					if( data.status == 'true' ) {
						$(form).find('.form-control').val('');
					}
					form_btn.prop('disabled', false).html(form_btn_old_msg);
					$(form_result_div).html(data.message).fadeIn('slow');
					setTimeout(function(){ $(form_result_div).fadeOut('slow') }, 6000);
				}
			});
		}
	});
})(jQuery);