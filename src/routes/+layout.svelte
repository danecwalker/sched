<script lang="ts">
  import { dev } from "$app/environment";
  import "../app.css";
  let { children } = $props();
</script>

{@render children()}
