<script lang="ts">
  import { cn } from "$lib/utils.ts";

  let {
    class: className,
    value = 0,
    max = 100,
  }: { class?: string; value?: number; max?: number } = $props();
</script>

<div
  data-slot="progress"
  class={cn("bg-primary/20 relative h-2 w-full overflow-hidden rounded-full", className)}
>
  <div
    data-slot="progress-indicator"
    class="bg-primary h-full w-full flex-1 transition-all"
    style="transform: translateX(-{100 - (100 * value) / max}%)"
  ></div>
</div>
