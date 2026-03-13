import type { Component } from 'solid-js';

const Unavailable: Component<{ label: string }> = (props) => {
  return (
    <div class="shell-unavailable">
      <wf-error-fallback
        title={`${props.label} is unavailable`}
        message="This service is not running or has no UI."
      />
    </div>
  );
};

export default Unavailable;
