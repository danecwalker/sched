import PocketBase from 'pocketbase';

const pb = new PocketBase('http://45.79.236.4:3002/');
pb.autoCancellation(false);
// option 1: authenticate as superuser using email/password (could be filled with ENV params)
await pb.collection('_superusers').authWithPassword("dane@danecwalker.com", "Charlie2017!", {
  // This will trigger auto refresh or auto reauthentication in case
  // the token has expired or is going to expire in the next 30 minutes.
  autoRefreshThreshold: 30 * 60
})

type ShiftCollection = {
  date: string,
  duration: number,
  break_duration: number,
  unpaid_break_duration: number,
  shifts: Shift[]
}

type Shift = {
  id: string,
  start_time: string,
  end_time: string,
  raw_start_time: Date,
  raw_end_time: Date,
  name: string,
  duration: number
}


const getShifts = async () => {
  console.log('getting shifts');
  const start = new Date();
  start.setDate(start.getDate() - start.getDay() + (start.getDay() === 0 ? -6 : 1));
  start.setHours(0, 0, 0, 0);
  const week_end = new Date(start);
  week_end.setDate(week_end.getDate() + 7);
  const end = new Date(start);
  end.setDate(end.getDate() + 14);


  let shifts = await pb.collection("shifts").getFullList({
    filter: `start_time > "${start.toISOString().replace("T", " ")}" && start_time < "${end.toISOString().replace("T", " ")}"`,
  });

  let blocks: {current: ShiftCollection[], next: ShiftCollection[]} = {
    current: [],
    next: []
  };

  let DateFormatter = new Intl.DateTimeFormat('en', {
    timeZone: 'Australia/Brisbane',
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  });

  const TimeFormatter = new Intl.DateTimeFormat('en', { 
    timeZone: 'Australia/Brisbane',
    hour: '2-digit', 
    minute: '2-digit', 
    hour12: false
  });

  for (let shift of shifts) {
    let date = new Date(shift.start_time);

    let start = new Date(shift.start_time);
    let end = new Date(shift.end_time);
    shift.duration = (end.getTime() - start.getTime()) / 1000 / 60 / 60;

    shift.raw_start_time = shift.start_time;
    shift.raw_end_time = shift.end_time
    shift.start_time = TimeFormatter.format(start);
    shift.end_time = TimeFormatter.format(end);

    if (date < week_end) {
      addShiftToBlock(blocks.current, shift, date, DateFormatter);
    } else {
      addShiftToBlock(blocks.next, shift, date, DateFormatter);
    }
  }

  durations(blocks.current);
  durations(blocks.next);

  return blocks
}

const durations = (blocks: ShiftCollection[]) => {
  for (let block of blocks) {
    block.duration = block.shifts.reduce((acc, shift) => acc + shift.duration, 0);
    if (block.duration < 4) {
      block.break_duration = 0;
      block.unpaid_break_duration = 0;
    } else if (block.duration < 5) {
      block.break_duration = 0.25;
      block.unpaid_break_duration = 0;
    } else if (block.duration < 7) {
      block.break_duration = 0.75;
      block.unpaid_break_duration = 0.5;
    } else if (block.duration < 10) {
      block.break_duration = 1;
      block.unpaid_break_duration = 0.5;
    } else {
      block.break_duration = 1.5;
      block.unpaid_break_duration = 1;
    }
  }
}

const addShiftToBlock = (blocks: ShiftCollection[], shift: any, date: Date, DateFormatter: Intl.DateTimeFormat) => {
  let existingBlock = blocks.find(b => b.date === DateFormatter.format(date));
  if (existingBlock) {
    existingBlock.shifts.push({
      id: shift.id,
      start_time: shift.start_time,
      end_time: shift.end_time,
      raw_start_time: shift.raw_start_time,
      raw_end_time: shift.raw_end_time,
      name: shift.name,
      duration: shift.duration
    });
  } else {
    blocks.push({
      date: DateFormatter.format(date),
      duration: 0,
      break_duration: 0,
      unpaid_break_duration: 0,
      shifts: [{
        id: shift.id,
        start_time: shift.start_time,
        end_time: shift.end_time,
        raw_start_time: shift.raw_start_time,
        raw_end_time: shift.raw_end_time,
        name: shift.name,
        duration: shift.duration
      }]
    });
  }
}

export {
  type ShiftCollection,
  type Shift,
  getShifts,
  pb
}
