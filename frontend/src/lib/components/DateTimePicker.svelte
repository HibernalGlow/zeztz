<script lang="ts">
	import Calendar from "$lib/components/ui/calendar/calendar.svelte";
	import * as Popover from "$lib/components/ui/popover/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import ChevronDownIcon from "@lucide/svelte/icons/chevron-down";
	import { getLocalTimeZone, CalendarDate, type DateValue } from "@internationalized/date";

	interface Props {
		value?: string;
		onchange?: (value: string) => void;
		id?: string;
	}

	let { value = $bindable(), onchange, id = "datetime" }: Props = $props();

	let open = $state(false);
	let dateValue = $state<CalendarDate | undefined>();
	let timeValue = $state("10:30:00");

	$effect(() => {
		if (value) {
			const [datePart, timePart] = value.split(" ");
			if (datePart) {
				const [year, month, day] = datePart.split("-").map(Number);
				dateValue = new CalendarDate(year, month, day);
			}
			if (timePart) {
				timeValue = timePart;
			}
		}
	});

	function updateValue() {
		if (dateValue) {
			const dateStr = dateValue.toString();
			const newValue = `${dateStr} ${timeValue}`;
			value = newValue;
			onchange?.(newValue);
		}
	}

	function onDateChange(v: DateValue | undefined) {
		dateValue = v as CalendarDate | undefined;
		open = false;
		updateValue();
	}

	function onTimeChange(e: Event) {
		timeValue = (e.target as HTMLInputElement).value;
		updateValue();
	}

	const displayDate = $derived(
		dateValue ? dateValue.toDate(getLocalTimeZone()).toLocaleDateString() : "选择日期"
	);
</script>

<div class="flex gap-2">
	<Popover.Root bind:open>
		<Popover.Trigger id="{id}-date">
			{#snippet child({ props })}
				<Button
					{...props}
					variant="outline"
					class="w-32 justify-between font-normal text-xs"
				>
					{displayDate}
					<ChevronDownIcon class="size-3" />
				</Button>
			{/snippet}
		</Popover.Trigger>
		<Popover.Content class="w-auto overflow-hidden p-0" align="start">
			<Calendar
				type="single"
				bind:value={dateValue}
				onValueChange={onDateChange}
				captionLayout="dropdown"
			/>
		</Popover.Content>
	</Popover.Root>
	<Input
		type="time"
		id="{id}-time"
		step="1"
		value={timeValue}
		onchange={onTimeChange}
		class="bg-background w-24 text-xs appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
	/>
</div>
