<script lang="ts">
  import { cn } from "$lib/utils.ts";

  let {
    class: className,
    children,
    ...restProps
  }: { class?: string; children?: import("svelte").Snippet } & Record<string, any> = $props();
</script>

<label
  data-slot="label"
  class={cn(
    "flex select-none items-center gap-2 text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-50",
    className
  )}
  {...restProps}
>
  {@render children?.()}
</label>
