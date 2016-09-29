Zepto(function(){
  var platformMappings = {
    "Mac": ".mac",
    "X11": ".lin",
    "Linux": ".lin",
    "Win": ".win"
  }
  var plat = navigator.platform
  var downloadToHighlight = ""
  $.each(platformMappings, function(key, value){
    if(plat.indexOf(key) != -1) {
      downloadToHighlight = value;
      return false
    }
  });

  if(downloadToHighlight !== ""){
    download = $(downloadToHighlight).first.find('.btn');
    download.removeClass("btn-secondary")
    download.addClass("btn-primary")
  }
})
