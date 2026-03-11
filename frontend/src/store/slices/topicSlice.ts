import type { SliceCreator } from '../types';
import { flattenActions } from '../utils';

export type Topic = {
  id: string;
  title: string;
  content: string;
  createdAt: number;
  updatedAt: number;
};

export type TopicState = {
  activeTopicId: string | null;
  topicMaps: Record<string, Topic>;
  topicIds: string[];
  topicsInit: boolean;
  topicLoadingIds: Set<string>;
};

export type TopicAction = Pick<TopicActionImpl, keyof TopicActionImpl>;
export type TopicStore = TopicState & TopicAction;

const _initialState: TopicState = {
  activeTopicId: null,
  topicMaps: {},
  topicIds: [],
  topicsInit: false,
  topicLoadingIds: new Set(),
};

type SetState = (
  partial:
    | TopicStore
    | Partial<TopicStore>
    | ((state: TopicStore) => TopicStore | Partial<TopicStore>),
  replace?: boolean,
) => void;

export const createTopicSlice = (
  set: SetState,
  get: () => TopicStore,
  _api?: unknown,
) => new TopicActionImpl(set, get, _api);

export class TopicActionImpl {
  readonly #set: SetState;
  readonly #get: () => TopicStore;

  constructor(set: SetState, get: () => TopicStore, _api?: unknown) {
    void _api;
    this.#set = set;
    this.#get = get;
  }

  // ==================== Public Actions ====================

  createTopic = async (params: { title: string; content: string }) => {
    if (!params.title.trim()) {
      throw new Error('Title is required');
    }
    return this.internal_createTopic(params);
  };

  updateTopic = async (
    id: string,
    params: Partial<Pick<Topic, 'title' | 'content'>>,
  ) => {
    const state = this.#get();
    if (!state.topicMaps[id]) {
      throw new Error(`Topic ${id} not found`);
    }
    return this.internal_updateTopic(id, params);
  };

  deleteTopic = async (id: string) => {
    const state = this.#get();
    if (!state.topicMaps[id]) {
      throw new Error(`Topic ${id} not found`);
    }
    return this.internal_deleteTopic(id);
  };

  switchTopic = (id: string | null) => {
    const state = this.#get();
    if (id && !state.topicMaps[id]) {
      console.warn(`Topic ${id} not found`);
      return;
    }
    this.#set({ activeTopicId: id });
  };

  useFetchTopics = async () => {
    if (this.#get().topicsInit) return;

    this.#set({ topicsInit: true });

    try {
      // TODO: Replace with actual service call
      // const topics = await topicService.fetchAll();
      // this.#dispatchTopic({ type: 'setTopics', value: topics });
    } catch (error) {
      this.#set({ topicsInit: false });
      throw error;
    }
  };

  // ==================== Internal Actions ====================

  internal_createTopic = async (params: { title: string; content: string }) => {
    const tmpId = `tmp_${Date.now()}`;
    const now = Date.now();

    const tempTopic: Topic = {
      id: tmpId,
      title: params.title,
      content: params.content,
      createdAt: now,
      updatedAt: now,
    };

    // 1. Optimistic update
    this.#dispatchTopic({ type: 'addTopic', value: tempTopic });

    this.#set((state) => ({
      topicLoadingIds: new Set(state.topicLoadingIds).add(tmpId),
    }));

    try {
      // 2. Call backend service
      // TODO: Replace with actual service call
      // const topicId = await topicService.create(params);

      // 3. Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 500));

      // 4. Mark as non-temporary (in real implementation, replace tmpId with real id)
      this.#set((state) => {
        const newLoadingIds = new Set(state.topicLoadingIds);
        newLoadingIds.delete(tmpId);
        return { topicLoadingIds: newLoadingIds };
      });

      return tmpId;
    } catch (error) {
      // 5. Rollback on error
      this.#dispatchTopic({ type: 'removeTopic', id: tmpId });
      this.#set((state) => {
        const newLoadingIds = new Set(state.topicLoadingIds);
        newLoadingIds.delete(tmpId);
        return { topicLoadingIds: newLoadingIds };
      });
      throw error;
    }
  };

  internal_updateTopic = async (
    id: string,
    params: Partial<Pick<Topic, 'title' | 'content'>>,
  ) => {
    const state = this.#get();
    const prevTopic = state.topicMaps[id];

    if (!prevTopic) {
      throw new Error(`Topic ${id} not found`);
    }

    // 1. Optimistic update
    this.#dispatchTopic({
      type: 'updateTopic',
      id,
      value: { ...prevTopic, ...params, updatedAt: Date.now() },
    });

    try {
      // 2. Call backend service
      // TODO: Replace with actual service call
      // await topicService.update(id, params);

      // 3. Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 300));
    } catch (error) {
      // 4. Rollback on error
      this.#dispatchTopic({ type: 'updateTopic', id, value: prevTopic });
      throw error;
    }
  };

  internal_deleteTopic = async (id: string) => {
    // Note: Don't use optimistic update for delete operations
    // The risk of destructive operations requires confirm-first approach

    try {
      // 1. Call backend service first
      // TODO: Replace with actual service call
      // await topicService.delete(id);

      // 2. Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 200));

      // 3. Update state after successful deletion
      this.#dispatchTopic({ type: 'removeTopic', id });
    } catch (error) {
      // No need to rollback, state is clean
      throw error;
    }
  };

  // ==================== Dispatch Methods ====================

  #dispatchTopic = (action: TopicActionType) => {
    this.#set((state) => topicReducer(state, action));
  };
}

// ==================== Reducer ====================

type TopicActionType =
  | { type: 'addTopic'; value: Topic }
  | { type: 'updateTopic'; id: string; value: Topic }
  | { type: 'removeTopic'; id: string }
  | { type: 'setTopics'; value: Topic[] };

function topicReducer(state: TopicState, action: TopicActionType): TopicState {
  switch (action.type) {
    case 'addTopic': {
      const topic = action.value;
      return {
        ...state,
        topicMaps: { ...state.topicMaps, [topic.id]: topic },
        topicIds: [...state.topicIds, topic.id],
      };
    }

    case 'updateTopic': {
      const { id, value } = action;
      if (!state.topicMaps[id]) return state;
      return {
        ...state,
        topicMaps: {
          ...state.topicMaps,
          [id]: value,
        },
      };
    }

    case 'removeTopic': {
      const { id } = action;
      const { [id]: _, ...remainingMaps } = state.topicMaps;
      return {
        ...state,
        topicMaps: remainingMaps,
        topicIds: state.topicIds.filter((tid) => tid !== id),
        activeTopicId: state.activeTopicId === id ? null : state.activeTopicId,
      };
    }

    case 'setTopics': {
      const topics = action.value;
      return {
        ...state,
        topicMaps: topics.reduce(
          (acc, topic) => {
            acc[topic.id] = topic;
            return acc;
          },
          {} as Record<string, Topic>,
        ),
        topicIds: topics.map((t) => t.id),
      };
    }

    default:
      return state;
  }
}

// ==================== Slice Export ====================

export const topicSlice: SliceCreator<TopicStore> = (...params) => ({
  ..._initialState,
  ...flattenActions<TopicAction>([
    createTopicSlice(params[0] as SetState, params[1], params[2]),
  ]),
});
