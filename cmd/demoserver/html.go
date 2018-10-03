package main

const html = `
<!doctype html>
<html>
	<body>
		<p>Events:</p>
		<div id="events"></div>

		<script>
			var events = document.getElementById('events');

			var source = new EventSource('/events');
			source.addEventListener('open', function (e) {
				console.log('open:', e);
			});
			source.addEventListener('error', function (e) {
				console.log('error:', e);
			});
			source.addEventListener('message', function (e) {
				console.log('message:', e.data);
				append(e.data);
			});
			source.addEventListener('urgentupdate', function (e) {
				console.log('urgent update:', e.data);
				var p = append(e.data)
				p.style.color = 'red';
			});

			function append(msg) {
				var p = document.createElement('p');
				p.textContent = msg;

				events.insertBefore(p, events.firstChild);
				return p
			}
		</script>
	</body>
</html>
`
